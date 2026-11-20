package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	// "slices"
	// "strings"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type MutexCache[T any] struct {
	inner T
	mutex sync.Mutex
}

type ElementoClient struct {
	Target            string
	Username          string
	Password          string
	Portal            string
	httpClient        *http.Client
	// These locks are always locked in the order cachedUUID/targetData -> cachedToken.
	// Avoid inverting the lock order anywhere to avoid introducing deadlocks.
	cachedToken       MutexCache[struct {
		token *string
		expiry int64
	}]
	cachedUUID        MutexCache[*string]
	targetData        MutexCache[*TargetGrant]
}

func NewElementoClient(target, username, password, portal string) *ElementoClient {
    // httpClient := &http.Client{
    //     Transport: &http.Transport{
    //         TLSClientConfig: &tls.Config{
    //             InsecureSkipVerify: true,
    //         },
    //     },
    // }

    return &ElementoClient{
		Target:     target,
		Username: username,
		Password: password,
		Portal:   portal,
		httpClient: http.DefaultClient,
	}
}

func (c *ElementoClient) request(
	ctx context.Context,
	method,
	url string,
	headers map[string]string,
	body any,
) (*http.Response, error) {
	tflog.Debug(ctx, fmt.Sprintf("Sending %s request to %s", method, url))

	var reqBody io.Reader = nil
	if body != nil {
		body, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}

		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	for header := range headers {
		req.Header.Set(header, headers[header])
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// closeResponseBody drains and closes an HTTP response body so the
// underlying connection can be reused by the client transport.
func closeResponseBody(resp *http.Response) {
    if resp == nil || resp.Body == nil {
        return
    }
    _, _ = io.Copy(io.Discard, resp.Body)
    _ = resp.Body.Close()
}

func (c *ElementoClient) getAuth(ctx context.Context) (*string, error) {
	// TODO: use golang.org/x/sync/singleflight instead of mutex locking
	c.cachedToken.mutex.Lock()
	defer c.cachedToken.mutex.Unlock()

	if c.cachedToken.inner.token != nil && c.cachedToken.inner.expiry - 60 * 2 > time.Now().Unix() {
		return c.cachedToken.inner.token, nil
	}

	url := "https://" + c.Portal + "/api/v2/auth/userpass"

	authRequest := authRequest{
		Username: c.Username,
		Password: c.Password,
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := c.request(ctx, "POST", url, headers, authRequest)
	if err != nil {
		return nil, err
	}

    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to authenticate: %s %s", resp.Status, string(body))
	}

	var authResponse authResponse
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	if err != nil {
		return nil, err
	}

	c.cachedToken.inner.token = &authResponse.AccessToken
	c.cachedToken.inner.expiry = authResponse.AccessTokenExp
	return c.cachedToken.inner.token, nil
}

func (c *ElementoClient) getClientUUID(ctx context.Context) (*string, error) {
	c.cachedUUID.mutex.Lock()
	defer c.cachedUUID.mutex.Unlock()

	if c.cachedUUID.inner != nil {
		return c.cachedUUID.inner, nil
	}

	authToken, err := c.getAuth(ctx)
	if err != nil {
		return nil, err
	}

    token, _, err := jwt.NewParser().ParseUnverified(*authToken, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, fmt.Errorf("invalid claims")
    }

    clientUUID, ok := claims["client_uuid"].(string)
    if !ok {
        return nil, fmt.Errorf("client_uuid claim missing or not a string")
    }

    c.cachedUUID.inner = &clientUUID

	return c.cachedUUID.inner, nil
}

func (c *ElementoClient) getTargetData(ctx context.Context) (*TargetGrant, error) {
	c.targetData.mutex.Lock()
	defer c.targetData.mutex.Unlock()

	if c.targetData.inner != nil {
		return c.targetData.inner, nil
	}

	url := "https://" + c.Portal + "/api/v2/grants/me"

	authToken, err := c.getAuth(ctx)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Authorization": "Bearer " + *authToken,
	}

	resp, err := c.request(ctx, "GET", url, headers, nil)
	if err != nil {
		return nil, err
	}

	defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get granted targets: %s %s", resp.Status, string(body))
	}

	var targetGrantsResponse TargetGrantsResponse
	err = json.NewDecoder(resp.Body).Decode(&targetGrantsResponse)
	if err != nil {
		return nil, err
	}

	var target *TargetGrant

	for _, grantedTarget := range targetGrantsResponse.Targets {
		if grantedTarget.TargetName == c.Target {
			target = &grantedTarget
			break
		}
	}

	if target == nil {
		return nil, fmt.Errorf("client does not have access to target %s", c.Target)
	}

	url = "https://" + c.Portal + "/api/v2/connections/me"
	resp, err = c.request(ctx, "GET", url, headers, nil)
	if err != nil {
		return nil, err
	}

	defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get active targets: %s %s", resp.Status, string(body))
	}

	var activeTargetsResponse ActiveTargetsResponse
	err = json.NewDecoder(resp.Body).Decode(&activeTargetsResponse)
	if err != nil {
		return nil, err
	}

	isActiveTarget := false
	for _, activeTarget := range activeTargetsResponse.Targets {
		if activeTarget.UUID == target.UUID {
			isActiveTarget = true
			break
		}
	}

	if !isActiveTarget {
		return nil, fmt.Errorf("target %s is not currently active", c.Target)
	}

	c.targetData.inner = target

	return target, nil
}

func (c *ElementoClient) getHost(ctx context.Context) (*string, error) {
	targetData, err := c.getTargetData(ctx)
	if err != nil {
		return nil, err
	}

	return &targetData.TargetConfig.IPList[0], nil
}

func (c *ElementoClient) getTargetUUID(ctx context.Context) (*string, error) {
	targetData, err := c.getTargetData(ctx)
	if err != nil {
		return nil, err
	}

	return &targetData.UUID, nil
}

func (c *ElementoClient) RenderCloudinit(ctx context.Context, service string, variables map[string]string) (*string, error) {
	authToken, err := c.getAuth(ctx)
	if err != nil {
		return nil, err
	}

	url := "https://spacedock.elemento.cloud/v1/marketplace/render"

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + *authToken,
	}

	request := renderCloudinitRequest{
		Service:   service,
		Variables: variables,
	}

	resp, err := c.request(ctx, "POST", url, headers, request)
	if err != nil {
		return nil, err
	}

	defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to render cloudinit: %s %s", resp.Status, string(body))
	}

	var response renderCloudinitResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response.Cloudinit, nil

}

func (c *ElementoClient) canCreate(ctx context.Context, resource any, method, urlSuffix, resourceName string, headers map[string]string) (*CanCreateResponse, error) {
	host, err := c.getHost(ctx)
	if err != nil {
		return nil, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth(ctx)
	if err != nil {
		return nil, err
	}

    target_uuid, err := c.getTargetUUID(ctx)
    if err != nil {
    	return nil, err
    }

    if headers == nil {
    	headers = map[string]string{}
    }

    headers["Content-Type"] = "application/json"
    headers["Authorization"] = "Bearer " + *authToken
    headers["X-Target-UUID"] = *target_uuid

    resp, err := c.request(ctx, method, url, headers, resource)
    if err != nil {
        return nil, err
    }

    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create %s resource: %s %s", resourceName, resp.Status, string(body))
	}

	var response CanCreateResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *ElementoClient) GetResourceStatus(ctx context.Context, billing_uuid string) (*string, error) {
	url := "https://" + c.Portal + "/api/v2/billing/" + billing_uuid + "/status"

	authToken, err := c.getAuth(ctx)
	if err != nil {
		return nil, err
	}
	headers := map[string]string{
		"Authorization": "Bearer " + *authToken,
	}

	resp, err := c.request(ctx, "GET", url, headers, nil)
	if err != nil {
		return nil, err
	}
    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("failed to get resource status: %s %s", resp.Status, string(body))
	}

	var statusResponse struct {
		Status string `json:"billing_status"`
	}
	err = json.NewDecoder(resp.Body).Decode(&statusResponse)
	if err != nil {
		return nil, err
	}

	return &statusResponse.Status, nil
}

func (c *ElementoClient) createResource(ctx context.Context, resource any, urlSuffix, resourceName string, headers map[string]string) (CreateResponse, error) {
	host, err := c.getHost(ctx)
	if err != nil {
		return nil, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth(ctx)
	if err != nil {
		return nil, err
	}

    target_uuid, err := c.getTargetUUID(ctx)
    if err != nil {
    	return nil, err
    }

    if headers == nil {
    	headers = map[string]string{}
    }
    headers["Content-Type"] = "application/json"
    headers["Authorization"] = "Bearer " + *authToken
    headers["X-Target-UUID"] = *target_uuid

    resp, err := c.request(ctx, "POST", url, headers, resource)
    if err != nil {
        return nil, err
    }

    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create %s resource: %s %s", resourceName, resp.Status, string(body))
	}

    targetData, err := c.getTargetData(ctx)
    if err != nil {
    	return nil, err
    }

    switch targetData.TargetType {
    	case "meson_private":
	        var response PrivateCreateResponse
	        err = json.NewDecoder(resp.Body).Decode(&response)
	        if err != nil {
	            return nil, err
	        }
	        return response, nil
     	case "meson_public":
			var response PublicCreateResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
			    return nil, err
			}
			return response, nil
		default:
			return nil, fmt.Errorf("unexpected status code %d when creating %s resource", resp.StatusCode, resourceName)
    }
}

func (c *ElementoClient) deleteResource(ctx context.Context, deleteRequest any, method, urlSuffix, resourceName string, headers map[string]string) (int, error) {
	host, err := c.getHost(ctx)
	if err != nil {
		return -1, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth(ctx)
	if err != nil {
		return -1, err
	}

    target_uuid, err := c.getTargetUUID(ctx)
    if err != nil {
    	return -1, err
    }

    if headers == nil {
    	headers = map[string]string{}
    }
    headers["Content-Type"] = "application/json"
    headers["Authorization"] = "Bearer " + *authToken
    headers["X-Target-UUID"] = *target_uuid

    resp, err := c.request(ctx, method, url, headers, deleteRequest)
    if err != nil {
        return -1, err
    }

    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, fmt.Errorf("failed to delete %s resource: %s %s", resourceName, resp.Status, string(body))
	}

	return resp.StatusCode, nil
}

func listResources[T any](c *ElementoClient, ctx context.Context, listRequest any, method, urlSuffix, resourceName string, headers map[string]string) ([]T, error) {
	host, err := c.getHost(ctx)
	if err != nil {
		return nil, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth(ctx)
	if err != nil {
		return nil, err
	}

    target_uuid, err := c.getTargetUUID(ctx)
    if err != nil {
    	return nil, err
    }

    if headers == nil {
    	headers = map[string]string{}
    }
    headers["Authorization"] = "Bearer " + *authToken
    headers["X-Target-UUID"] = *target_uuid

    if listRequest != nil {
        headers["Content-Type"] = "application/json"
    }

	resp, err := c.request(ctx, method, url, headers, listRequest)
	if err != nil {
		return nil, err
	}

    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list %s resources: %s %s", resourceName, resp.Status, string(body))
	}

	var resources []T

	err = json.NewDecoder(resp.Body).Decode(&resources)
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func getResource[T any](c *ElementoClient, ctx context.Context, getRequest any, method, urlSuffix, resourceName string, headers map[string]string) (*T, int, error) {
	host, err := c.getHost(ctx)
	if err != nil {
		return nil, -1, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth(ctx)
	if err != nil {
		return nil, -1, err
	}

    target_uuid, err := c.getTargetUUID(ctx)
    if err != nil {
    	return nil, -1, err
    }

    if headers == nil {
    	headers = map[string]string{}
    }
    headers["Authorization"] = "Bearer " + *authToken
    headers["X-Target-UUID"] = *target_uuid

    if getRequest != nil {
        headers["Content-Type"] = "application/json"
    }

	resp, err := c.request(ctx, method, url, headers, getRequest)
	if err != nil {
		return nil, -1, err
	}

    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, resp.StatusCode, fmt.Errorf("failed to get %s resource: %s %s", resourceName, resp.Status, string(body))
	}

	var resource T

	err = json.NewDecoder(resp.Body).Decode(&resource)
	if err != nil {
		return nil, -1, err
	}

	return &resource, resp.StatusCode, nil
}

type deserializer[T any] func(io.Reader) (*T, error)

func getCredentials[T any](c *ElementoClient, ctx context.Context, urlSuffix, resourceName string, deserialize deserializer[T], headers map[string]string) (*T, error) {
	host, err := c.getHost(ctx)
	if err != nil {
		return nil, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth(ctx)
	if err != nil {
		return nil, err
	}

    target_uuid, err := c.getTargetUUID(ctx)
    if err != nil {
    	return nil, err
    }

    if headers == nil {
    	headers = map[string]string{}
    }
    headers["Authorization"] = "Bearer " + *authToken
    headers["X-Target-UUID"] = *target_uuid

	resp, err := c.request(ctx, "GET", url, headers, nil)
	if err != nil {
		return nil, err
	}

    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get %s credentials: %s %s", resourceName, resp.Status, string(body))
	}

	resource, err := deserialize(resp.Body)
	if err != nil {
		return nil, err
	}

	return resource, nil
}


func (c *ElementoClient) CanCreateComputeResource(ctx context.Context, computeResource ComputeResource) (*CanCreateResponse, error) {
	targetData, err := c.getTargetData(ctx)
	if err != nil {
		return nil, err
	}

	var resource any
	var headers map[string]string
	if targetData.TargetType == "meson_public" || targetData.TargetType == "meson_private" {
		cloudinit, err := c.RenderCloudinit(ctx, computeResource.CloudInit.Template, computeResource.CloudInit.Variables)
		if err != nil {
			return nil, err
		}
		resource = computeResource.toSpotVM(cloudinit)
		headers = map[string]string{"X-Region": computeResource.Region}
	} else {
		resource = computeResource
	}

    return c.canCreate(ctx, resource, "GET", ":7777/api/v1.0/canallocate", "compute", headers)
}

func (c *ElementoClient) CreateComputeResource(ctx context.Context, computeResource ComputeResource) (CreateResponse, error) {
	targetData, err := c.getTargetData(ctx)
	if err != nil {
		return nil, err
	}

	var resource any
	var headers map[string]string
	if targetData.TargetType == "meson_public" || targetData.TargetType == "meson_private" {
		cloudinit, err := c.RenderCloudinit(ctx, computeResource.CloudInit.Template, computeResource.CloudInit.Variables)
		if err != nil {
			return nil, err
		}
		resource = computeResource.toSpotVM(cloudinit)
		headers = map[string]string{"X-Region": computeResource.Region}
	} else {
		resource = computeResource
	}

    return c.createResource(ctx, resource, ":7777/api/v1.0/register", "compute", headers)
}

func (c *ElementoClient) ListComputeResources(ctx context.Context) ([]ComputeListItem, error) {
	targetData, err := c.getTargetData(ctx)
	if err != nil {
		return nil, err
	}

	host, err := c.getHost(ctx)
	if err != nil {
		return nil, err
	}

	// Legacy compute handling needs special handling as the return payload
	// is in the form of {VMS: [VMItem]} rather than just [VMItem]
    url := "https://" + *host + ":7777/api/v1.0/running"

    authToken, err := c.getAuth(ctx)
    if err != nil {
        return nil, err
    }

    target_uuid, err := c.getTargetUUID(ctx)
    if err != nil {
    	return nil, err
    }

    headers := map[string]string{
        "Authorization": "Bearer " + *authToken,
        "X-Target-UUID": *target_uuid,
    }

	resp, err := c.request(ctx, "GET", url, headers, nil)
	if err != nil {
		return nil, err
	}

    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list compute resources: %s %s", resp.Status, string(body))
	}

	if targetData.TargetType == "meson_public" || targetData.TargetType == "meson_private" {
		var computeResources SpotVMListResponse
	    err = json.NewDecoder(resp.Body).Decode(&computeResources)
	    if err != nil {
	        return nil, err
	    }

	    return computeResources.toComputeListResponse().VMS, nil
	}

    var computeResources ComputeListResponse
    err = json.NewDecoder(resp.Body).Decode(&computeResources)
    if err != nil {
        return nil, err
    }

	return computeResources.VMS, nil
}

func (c *ElementoClient) DeleteComputeResource(ctx context.Context, uuid string) (int, error) {
    delRequest := struct {
        UUID string `json:"vm_uid"`
    }{
        UUID: uuid,
    }
    return c.deleteResource(ctx, delRequest, "DELETE", ":7777/api/v1.0/unregister", "compute", nil)
}


func (c *ElementoClient) CanCreateStorageResource(ctx context.Context, req CanCreateStorageRequest) (*CanCreateResponse, error) {
    return c.canCreate(ctx, req, "GET", ":7772/api/v1.0/cancreate", "storage", nil)
}

func (c *ElementoClient) CreateStorageResource(ctx context.Context, storageResource *StorageResource) (CreateResponse, error) {
    uuid, err := c.getClientUUID(ctx)
    if err != nil {
        return nil, err
    }
    storageResource.CreatorUUID = *uuid

    return c.createResource(ctx, storageResource, ":7772/api/v1.0/create", "storage", nil)
}

func (c *ElementoClient) DeleteStorageResource(ctx context.Context, volume_uuid string) (int, error) {
	client_uuid, err := c.getClientUUID(ctx)
	if err != nil {
		return -1, err
	}

	delRequest := struct {
		UUID       string `json:"volume_id"`
		ClientUUID string `json:"client_id"`
	}{
		UUID:       volume_uuid,
		ClientUUID: *client_uuid,
	}

    return c.deleteResource(ctx, delRequest, "POST", ":7772/api/v1.0/destroy", "storage", nil)
}

func (c *ElementoClient) ListStorageResources(ctx context.Context) ([]StorageResourceResponse, error) {
	client_uuid, err := c.getClientUUID(ctx)
	if err != nil {
		return nil, err
	}

	var listRequest = struct {
		ClientUUID string `json:"client_uid"`
	}{
		ClientUUID: *client_uuid,
	}

    return listResources[StorageResourceResponse](c, ctx, listRequest, "POST", ":7772/api/v1.0/accessible", "storage", nil)
}

func (c *ElementoClient) GetStorageResource(ctx context.Context, volume_uuid string) (*StorageResourceResponse, int, error) {
	client_uuid, err := c.getClientUUID(ctx)
	if err != nil {
		return nil, -1, err
	}

	var getRequest = struct {
		ClientUUID string `json:"client_uid"`
		VolumeUUID string `json:"volume_id"`
	}{
		ClientUUID: *client_uuid,
		VolumeUUID: volume_uuid,
	}

    return getResource[StorageResourceResponse](c, ctx, getRequest, "POST", ":7772/api/v1.0/info", "storage", nil)
}


func (c *ElementoClient) CanCreateObjectStorageResource(ctx context.Context, objectStorageResource ObjectStorageRequest) (*CanCreateResponse, error) {
    return c.canCreate(ctx, objectStorageResource, "POST", ":7770/api/v1.0/objectstorage/cancreate", "objectstorage", nil)
}

func (c *ElementoClient) CreateObjectStorageResource(ctx context.Context, objectStorageResource ObjectStorageRequest) (CreateResponse, error) {
    return c.createResource(ctx, objectStorageResource, ":7770/api/v1.0/objectstorage/create", "objectstorage", nil)
}

func (c *ElementoClient) DeleteObjectStorageResource(ctx context.Context, service_uuid string) (int, error) {
	delRequest := struct {
		UUID string `json:"service_uuid"`
	}{
		UUID: service_uuid,
	}

    return c.deleteResource(ctx, delRequest, "DELETE", ":7770/api/v1.0/objectstorage/delete", "objectstorage", nil)
}

func (c *ElementoClient) ListObjectStorageResources(ctx context.Context) ([]ObjectStorageResponse, error) {
    return listResources[ObjectStorageResponse](c, ctx, nil, "GET", ":7770/api/v1.0/objectstorage/running", "objectstorage", nil)
}

func (c *ElementoClient) GetObjectStorageResource(ctx context.Context, service_uuid string) (*ObjectStorageResponse, int, error) {
    return getResource[ObjectStorageResponse](c, ctx, nil, "GET", ":7770/api/v1.0/objectstorage/running/" + service_uuid, "objectstorage", nil)
}

func (c *ElementoClient) GetObjectStorageCredentials(ctx context.Context, service_uuid string) (*ObjectStorageCredentials, error) {
	urlSuffix := ":7770/api/v1.0/objectstorage/credentials/" + service_uuid

	serializer := func(r io.Reader) (*ObjectStorageCredentials, error) {
		return KVDeserializer[ObjectStorageCredentials](r, "=")
	}

    return getCredentials(c, ctx, urlSuffix, "objectstorage", serializer, nil)
}


func (c *ElementoClient) CanCreateDBAASResource(ctx context.Context, dbaasResource DBAASRequest) (*CanCreateResponse, error) {
    return c.canCreate(ctx, dbaasResource, "POST", ":7770/api/v1.0/dbaas/cancreate", "dbaas", nil)
}

func (c *ElementoClient) CreateDBAASResource(ctx context.Context, dbaasResource DBAASRequest) (CreateResponse, error) {
    return c.createResource(ctx, dbaasResource, ":7770/api/v1.0/dbaas/create", "dbaas", nil)
}

func (c *ElementoClient) DeleteDBAASResource(ctx context.Context, service_uuid string) (int, error) {
	delRequest := struct {
		UUID string `json:"service_uuid"`
	}{
		UUID: service_uuid,
	}

    return c.deleteResource(ctx, delRequest, "DELETE", ":7770/api/v1.0/dbaas/delete", "dbaas", nil)
}

func (c *ElementoClient) ListDBAASResources(ctx context.Context) ([]DBAASResponse, error) {
    return listResources[DBAASResponse](c, ctx, nil, "GET", ":7770/api/v1.0/dbaas/running", "dbaas", nil)
}

func (c *ElementoClient) GetDBAASResource(ctx context.Context, service_uuid string) (*DBAASResponse, int, error) {
    return getResource[DBAASResponse](c, ctx, nil, "GET", ":7770/api/v1.0/dbaas/running/" + service_uuid, "dbaas", nil)
}

func (c *ElementoClient) GetDBAASCredentials(ctx context.Context, service_uuid string) (*DBAASCredentials, error) {
	urlSuffix := ":7770/api/v1.0/dbaas/credentials/" + service_uuid

	serializer := func(r io.Reader) (*DBAASCredentials, error) {
		return KVDeserializer[DBAASCredentials](r, "=")
	}

    return getCredentials(c, ctx, urlSuffix, "dbaas", serializer, nil)
}


func (c *ElementoClient) CanCreateKAASResource(ctx context.Context, kaasResource KAASRequest) (*CanCreateResponse, error) {
    return c.canCreate(ctx, kaasResource, "POST", ":7770/api/v1.0/kaas/cancreate", "kaas", nil)
}

func (c *ElementoClient) CreateKAASResource(ctx context.Context, kaasResource KAASRequest) (CreateResponse, error) {
    return c.createResource(ctx, kaasResource, ":7770/api/v1.0/kaas/create", "kaas", nil)
}

func (c *ElementoClient) DeleteKAASResource(ctx context.Context, service_uuid string) (int, error) {
	delRequest := struct {
		UUID string `json:"service_uuid"`
	}{
		UUID: service_uuid,
	}

    return c.deleteResource(ctx, delRequest, "DELETE", ":7770/api/v1.0/kaas/delete", "kaas", nil)
}

func (c *ElementoClient) ListKAASResources(ctx context.Context) ([]KAASResponse, error) {
    return listResources[KAASResponse](c, ctx, nil, "GET", ":7770/api/v1.0/kaas/running", "kaas", nil)
}

func (c *ElementoClient) GetKAASResource(ctx context.Context, service_uuid string) (*KAASResponse, int, error) {
    return getResource[KAASResponse](c, ctx, nil, "GET", ":7770/api/v1.0/kaas/running/" + service_uuid, "kaas", nil)
}

func (c *ElementoClient) GetKAASCredentials(ctx context.Context, service_uuid string) (*string, error) {
	urlSuffix := ":7770/api/v1.0/kaas/credentials/" + service_uuid

	deserializer := func(r io.Reader) (*string, error) {
		credentials, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

        creds := string(credentials)
        return &creds, nil
    }
    return getCredentials(c, ctx, urlSuffix, "kaas", deserializer, nil)
}


// func (c *ElementoClient) ListN8NResources() ([]ComputeListItem, error) {
//     ret, err := c.ListComputeResources(ctx)
//     if err != nil {
//         return nil, err
//     }

//     var resources []ComputeListItem
//     for _, item := range ret {
//         if slices.Contains(strings.Split(item.Tags, ":"), "n8n") {
//             resources = append(resources, item)
//         }
//     }

//     return resources, nil
// }

func (c *ElementoClient) GetN8NCredentials(ctx context.Context, service_uuid string) (*N8NCredentials, error) {
	urlSuffix := ":7777/api/v1.0/credentials/" + service_uuid

	serializer := func(r io.Reader) (*N8NCredentials, error) {
		return KVDeserializer[N8NCredentials](r, ": |=")
	}

    return getCredentials(c, ctx, urlSuffix, "n8n", serializer, nil)
}


// func (c *ElementoClient) ListOpenClawResources() ([]ComputeListItem, error) {
//     ret, err := c.ListComputeResources(ctx)
//     if err != nil {
//         return nil, err
//     }

//     var resources []ComputeListItem
//     for _, item := range ret {
//         if slices.Contains(strings.Split(item.Tags, ":"), "openclaw") {
//             resources = append(resources, item)
//         }
//     }

//     return resources, nil
// }

func (c *ElementoClient) GetOpenClawCredentials(ctx context.Context, service_uuid string) (*OpenClawCredentials, error) {
	urlSuffix := ":7777/api/v1.0/credentials/" + service_uuid

	serializer := func(r io.Reader) (*OpenClawCredentials, error) {
		return KVDeserializer[OpenClawCredentials](r, ": |=")
	}

    return getCredentials(c, ctx, urlSuffix, "openclaw", serializer, nil)
}


func (c *ElementoClient) CanCreateFileStorageResource(ctx context.Context, fileStorageResource FileStorageRequest) (*CanCreateResponse, error) {
    return c.canCreate(ctx, fileStorageResource, "POST", ":7770/api/v1.0/filestorage/cancreate", "filestorage", nil)
}

func (c *ElementoClient) CreateFileStorageResource(ctx context.Context, fileStorageResource FileStorageRequest) (CreateResponse, error) {
    return c.createResource(ctx, fileStorageResource, ":7770/api/v1.0/filestorage/create", "filestorage", nil)
}

func (c *ElementoClient) DeleteFileStorageResource(ctx context.Context, service_uuid string) (int, error) {
    delRequest := struct {
        UUID       string `json:"service_uuid"`
    }{
        UUID: service_uuid,
    }

    return c.deleteResource(ctx, delRequest, "DELETE", ":7770/api/v1.0/filestorage/delete", "filestorage", nil)
}

func (c *ElementoClient) ListFileStorageResources(ctx context.Context) ([]FileStorageResponse, error) {
    return listResources[FileStorageResponse](c, ctx, nil, "GET", ":7770/api/v1.0/filestorage/running", "filestorage", nil)
}

func (c *ElementoClient) GetFileStorageResource(ctx context.Context, service_uuid string) (*FileStorageResponse, int, error) {
    return getResource[FileStorageResponse](c, ctx, nil, "GET", ":7770/api/v1.0/filestorage/running/" + service_uuid, "filestorage", nil)
}

func (c *ElementoClient) GetFileStorageCredentials(ctx context.Context, service_uuid string) (*FileStorageCredentials, error) {
    urlSuffix := ":7770/api/v1.0/filestorage/credentials/" + service_uuid

    serializer := func (r io.Reader) (*FileStorageCredentials, error) {
        return KVDeserializer[FileStorageCredentials](r, "=")
    }

    return getCredentials(c, ctx, urlSuffix, "filestorage", serializer, nil)
}


func (c *ElementoClient) CanCreatePrivateNetworkResource(ctx context.Context, privateNetworkResource PrivateNetworkRequest) (*CanCreateResponse, error) {
    return c.canCreate(ctx, privateNetworkResource, "POST", ":7770/api/v1.0/privatenetwork/cancreate", "privatenetwork", nil)
}

func (c *ElementoClient) CreatePrivateNetworkResource(ctx context.Context, privateNetworkResource PrivateNetworkRequest) (CreateResponse, error) {
    return c.createResource(ctx, privateNetworkResource, ":7770/api/v1.0/privatenetwork/create", "privatenetwork", nil)
}

func (c *ElementoClient) DeletePrivateNetworkResource(ctx context.Context, service_uuid string) (int, error) {
    delRequest := struct {
        UUID       string `json:"service_uuid"`
    }{
        UUID: service_uuid,
    }

    return c.deleteResource(ctx, delRequest, "DELETE", ":7770/api/v1.0/privatenetwork/delete", "privatenetwork", nil)
}

func (c *ElementoClient) ListPrivateNetworkResources(ctx context.Context) ([]PrivateNetworkResponse, error) {
    return listResources[PrivateNetworkResponse](c, ctx, nil, "GET", ":7770/api/v1.0/privatenetwork/running", "privatenetwork", nil)
}

func (c *ElementoClient) GetPrivateNetworkResource(ctx context.Context, service_uuid string) (*PrivateNetworkResponse, int, error) {
    return getResource[PrivateNetworkResponse](c, ctx, nil, "GET", ":7770/api/v1.0/privatenetwork/running/" + service_uuid, "privatenetwork", nil)
}

func (c *ElementoClient) GetPrivateNetworkCredentials(ctx context.Context, service_uuid string) (*PrivateNetworkCredentials, error) {
    urlSuffix := ":7770/api/v1.0/privatenetwork/credentials/" + service_uuid

    serializer := func (r io.Reader) (*PrivateNetworkCredentials, error) {
        return KVDeserializer[PrivateNetworkCredentials](r, "=")
    }

    return getCredentials(c, ctx, urlSuffix, "privatenetwork", serializer, nil)
}


func (c *ElementoClient) CanCreateResource(ctx context.Context, resource resourceWithCanCreate) (*CanCreateResponse, error) {
	cfg := resource.canCreateConfig()
    return c.canCreate(ctx, resource, cfg.Method, cfg.UrlSuffix, cfg.ResourceName, cfg.Headers)
}

func (c *ElementoClient) CreateResource(ctx context.Context, resource resourceWithCreate) (CreateResponse, error) {
	cfg := resource.createConfig()
    return c.createResource(ctx, resource, cfg.UrlSuffix, cfg.ResourceName, cfg.Headers)
}

func (c *ElementoClient) DeleteResource(ctx context.Context, service_uuid string, resourceName string) (int, error) {
    delRequest := struct {
        UUID       string `json:"service_uuid"`
    }{
        UUID: service_uuid,
    }

    return c.deleteResource(ctx, delRequest, "DELETE", fmt.Sprintf(":7770/api/v1.0/%s/delete", resourceName), resourceName, nil)
}

func ListResources[T any](c *ElementoClient, ctx context.Context, resourceName string) ([]T, error) {
    return listResources[T](c, ctx, nil, "GET", fmt.Sprintf(":7770/api/v1.0/%s/running", resourceName), resourceName, nil)
}

func GetResource[T any](c *ElementoClient, ctx context.Context, service_uuid string, resourceName string) (*T, int, error) {
    return getResource[T](c, ctx, nil, "GET", fmt.Sprintf(":7770/api/v1.0/%s/running/%s", resourceName, service_uuid), resourceName, nil)
}

func (c *ElementoClient) GetRegistryCredentials(ctx context.Context, service_uuid string) (*RegistryCredentials, error) {
    urlSuffix := ":7770/api/v1.0/registry/credentials/" + service_uuid

    serializer := func (r io.Reader) (*RegistryCredentials, error) {
        return KVDeserializer[RegistryCredentials](r, "=")
    }

    return getCredentials(c, ctx, urlSuffix, "registry", serializer, nil)
}

func (c *ElementoClient) GetServerlessCredentials(ctx context.Context, service_uuid string) (*ServerlessCredentials, error) {
    urlSuffix := ":7770/api/v1.0/serverless/credentials/" + service_uuid

    serializer := func (r io.Reader) (*ServerlessCredentials, error) {
        return KVDeserializer[ServerlessCredentials](r, "=")
    }

    return getCredentials(c, ctx, urlSuffix, "serverless", serializer, nil)
}

func (c *ElementoClient) GetAIServerlessCredentials(ctx context.Context, service_uuid string) (*AIServerlessCredentials, error) {
    urlSuffix := ":7770/api/v1.0/aiserverless/credentials/" + service_uuid

    serializer := func (r io.Reader) (*AIServerlessCredentials, error) {
        return KVDeserializer[AIServerlessCredentials](r, "=")
    }

    return getCredentials(c, ctx, urlSuffix, "aiserverless", serializer, nil)
}

func (c *ElementoClient) GetQueueAASCredentials(ctx context.Context, service_uuid string) (*QueueAASCredentials, error) {
    urlSuffix := ":7770/api/v1.0/queueaas/credentials/" + service_uuid

    serializer := func (r io.Reader) (*QueueAASCredentials, error) {
        return KVDeserializer[QueueAASCredentials](r, "=")
    }

    return getCredentials(c, ctx, urlSuffix, "queueaas", serializer, nil)
}


func (c *ElementoClient) ListSpotVMResources(ctx context.Context) ([]SpotVMListItem, error) {
	host, err := c.getHost(ctx)
	if err != nil {
		return nil, err
	}

	// Legacy compute handling needs special handling as the return payload
	// is in the form of {VMS: [VMItem]} rather than just [VMItem]
    url := "https://" + *host + ":7777/api/v1.0/running"

    authToken, err := c.getAuth(ctx)
    if err != nil {
        return nil, err
    }

    target_uuid, err := c.getTargetUUID(ctx)
    if err != nil {
    	return nil, err
    }

    headers := map[string]string{
        "Authorization": "Bearer " + *authToken,
        "X-Target-UUID": *target_uuid,
    }

    resp, err := c.request(ctx, "GET", url, headers, nil)
    if err != nil {
        return nil, err
    }

    defer closeResponseBody(resp)

    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("failed to list compute resources: %s %s", resp.Status, string(body))
    }

	var computeResources SpotVMListResponse
    err = json.NewDecoder(resp.Body).Decode(&computeResources)
    if err != nil {
        return nil, err
    }

    return computeResources.VMS, nil
}

func (c *ElementoClient) DeleteSpotVMResource(ctx context.Context, uuid string) (int, error) {
    delRequest := struct {
        UUID string `json:"vm_uid"`
    }{
        UUID: uuid,
    }
    return c.deleteResource(ctx, delRequest, "DELETE", ":7777/api/v1.0/unregister", "spotvm", nil)
}
