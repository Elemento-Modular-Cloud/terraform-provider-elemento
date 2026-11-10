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
	// "github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/golang-jwt/jwt/v5"
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
	context           context.Context
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

func NewElementoClient(ctx context.Context, target, username, password, portal string) *ElementoClient {
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
		context: ctx,
	}
}

func (c *ElementoClient) request(
	method,
	url string,
	headers map[string]string,
	body any,
) (*http.Response, error) {
	var reqBody io.Reader = nil
	if body != nil {
		body, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
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

func (c *ElementoClient) getAuth() (*string, error) {
	// TODO: use golang.org/x/sync/singleflight instead of mutex locking
	c.cachedToken.mutex.Lock()
	defer c.cachedToken.mutex.Unlock()

	if c.cachedToken.inner.token != nil && c.cachedToken.inner.expiry > time.Now().Unix() {
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

	resp, err := c.request("POST", url, headers, authRequest)
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

func (c *ElementoClient) getClientUUID() (*string, error) {
	c.cachedUUID.mutex.Lock()
	defer c.cachedUUID.mutex.Unlock()

	if c.cachedUUID.inner != nil {
		return c.cachedUUID.inner, nil
	}

	authToken, err := c.getAuth()
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

func (c *ElementoClient) getTargetData() (*TargetGrant, error) {
	c.targetData.mutex.Lock()
	defer c.targetData.mutex.Unlock()

	if c.targetData.inner != nil {
		return c.targetData.inner, nil
	}

	url := "https://" + c.Portal + "/api/v2/grants/me"

	authToken, err := c.getAuth()
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Authorization": "Bearer " + *authToken,
	}

	resp, err := c.request("GET", url, headers, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

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
	resp, err = c.request("GET", url, headers, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

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

func (c *ElementoClient) getHost() (*string, error) {
	targetData, err := c.getTargetData()
	if err != nil {
		return nil, err
	}

	return &targetData.TargetConfig.IPList[0], nil
}

func (c *ElementoClient) getTargetUUID() (*string, error) {
	targetData, err := c.getTargetData()
	if err != nil {
		return nil, err
	}

	return &targetData.UUID, nil
}

func (c *ElementoClient) RenderCloudinit(service string, variables map[string]string) (*string, error) {
	authToken, err := c.getAuth()
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

	resp, err := c.request("POST", url, headers, request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

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

func (c *ElementoClient) canCreate(resource any, method, urlSuffix, resourceName string, headers map[string]string) (*CanCreateResponse, error) {
	host, err := c.getHost()
	if err != nil {
		return nil, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth()
	if err != nil {
		return nil, err
	}

    target_uuid, err := c.getTargetUUID()
    if err != nil {
    	return nil, err
    }

    if headers == nil {
    	headers = map[string]string{}
    }

    headers["Content-Type"] = "application/json"
    headers["Authorization"] = "Bearer " + *authToken
    headers["X-Target-UUID"] = *target_uuid

    resp, err := c.request(method, url, headers, resource)
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

func (c *ElementoClient) GetResourceStatus(billing_uuid string) (*string, error) {
	url := "https://" + c.Portal + "/api/v2/billing/" + billing_uuid + "/status"

	authToken, err := c.getAuth()
	if err != nil {
		return nil, err
	}
	headers := map[string]string{
		"Authorization": "Bearer " + *authToken,
	}

	resp, err := c.request("GET", url, headers, nil)
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

func (c *ElementoClient) createResource(resource any, urlSuffix, resourceName string, headers map[string]string) (CreateResponse, error) {
	host, err := c.getHost()
	if err != nil {
		return nil, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth()
	if err != nil {
		return nil, err
	}

    target_uuid, err := c.getTargetUUID()
    if err != nil {
    	return nil, err
    }

    if headers == nil {
    	headers = map[string]string{}
    }
    headers["Content-Type"] = "application/json"
    headers["Authorization"] = "Bearer " + *authToken
    headers["X-Target-UUID"] = *target_uuid

    resp, err := c.request("POST", url, headers, resource)
    if err != nil {
        return nil, err
    }

    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create %s resource: %s %s", resourceName, resp.Status, string(body))
	}

    targetData, err := c.getTargetData()
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

func (c *ElementoClient) deleteResource(deleteRequest any, method, urlSuffix, resourceName string, headers map[string]string) (int, error) {
	host, err := c.getHost()
	if err != nil {
		return -1, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth()
	if err != nil {
		return, -1, err
	}

    target_uuid, err := c.getTargetUUID()
    if err != nil {
    	return -1, err
    }

    if headers == nil {
    	headers = map[string]string{}
    }
    headers["Content-Type"] = "application/json"
    headers["Authorization"] = "Bearer " + *authToken
    headers["X-Target-UUID"] = *target_uuid

    resp, err := c.request(method, url, headers, deleteRequest)
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

func listResources[T any](c *ElementoClient, listRequest any, method, urlSuffix, resourceName string, headers map[string]string) ([]T, error) {
	host, err := c.getHost()
	if err != nil {
		return nil, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth()
	if err != nil {
		return nil, err
	}

    target_uuid, err := c.getTargetUUID()
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

	resp, err := c.request(method, url, headers, listRequest)
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

func getResource[T any](c *ElementoClient, getRequest any, method, urlSuffix, resourceName string, headers map[string]string) (*T, int, error) {
	host, err := c.getHost()
	if err != nil {
		return nil, -1, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth()
	if err != nil {
		return nil, -1, err
	}

    target_uuid, err := c.getTargetUUID()
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

	resp, err := c.request(method, url, headers, getRequest)
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

func getCredentials[T any](c *ElementoClient, urlSuffix, resourceName string, deserialize deserializer[T], headers map[string]string) (*T, error) {
	host, err := c.getHost()
	if err != nil {
		return nil, err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth()
	if err != nil {
		return nil, err
	}

    target_uuid, err := c.getTargetUUID()
    if err != nil {
    	return nil, err
    }

    if headers == nil {
    	headers = map[string]string{}
    }
    headers["Authorization"] = "Bearer " + *authToken
    headers["X-Target-UUID"] = *target_uuid

	resp, err := c.request("GET", url, headers, nil)
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

func (c *ElementoClient) CanCreateComputeResource(computeResource ComputeResource) (*CanCreateResponse, error) {
	targetData, err := c.getTargetData()
	if err != nil {
		return nil, err
	}

	var resource any
	var headers map[string]string
	if targetData.TargetType == "meson_public" || targetData.TargetType == "meson_private" {
		cloudinit, err := c.RenderCloudinit(computeResource.CloudInit.Template, computeResource.CloudInit.Variables)
		if err != nil {
			return nil, err
		}
		resource = computeResource.toSpotVM(cloudinit)
		headers = map[string]string{"X-Region": computeResource.Region}
	} else {
		resource = computeResource
	}

    return c.canCreate(resource, "GET", ":7777/api/v1.0/canallocate", "compute", headers)
}

func (c *ElementoClient) CreateComputeResource(computeResource ComputeResource) (CreateResponse, error) {
	targetData, err := c.getTargetData()
	if err != nil {
		return nil, err
	}

	var resource any
	var headers map[string]string
	if targetData.TargetType == "meson_public" || targetData.TargetType == "meson_private" {
		cloudinit, err := c.RenderCloudinit(computeResource.CloudInit.Template, computeResource.CloudInit.Variables)
		if err != nil {
			return nil, err
		}
		resource = computeResource.toSpotVM(cloudinit)
		headers = map[string]string{"X-Region": computeResource.Region}
	} else {
		resource = computeResource
	}

    return c.createResource(resource, ":7777/api/v1.0/register", "compute", headers)
}

func (c *ElementoClient) ListComputeResources() ([]ComputeListItem, error) {
	targetData, err := c.getTargetData()
	if err != nil {
		return nil, err
	}

	host, err := c.getHost()
	if err != nil {
		return nil, err
	}

	// Legacy compute handling needs special handling as the return payload
	// is in the form of {VMS: [VMItem]} rather than just [VMItem]
    url := "https://" + *host + ":7777/api/v1.0/running"

    authToken, err := c.getAuth()
    if err != nil {
        return nil, err
    }

    target_uuid, err := c.getTargetUUID()
    if err != nil {
    	return nil, err
    }

    headers := map[string]string{
        "Authorization": "Bearer " + *authToken,
        "X-Target-UUID": *target_uuid,
    }

	resp, err := c.request("GET", url, headers, nil)
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

func (c *ElementoClient) DeleteComputeResource(uuid string) (int, error) {
    delRequest := struct {
        UUID string `json:"vm_uid"`
    }{
        UUID: uuid,
    }
    return c.deleteResource(delRequest, "DELETE", ":7777/api/v1.0/unregister", "compute", nil)
}

func (c *ElementoClient) CanCreateStorageResource(req CanCreateStorageRequest) (*CanCreateResponse, error) {
    return c.canCreate(req, "GET", ":7772/api/v1.0/cancreate", "storage", nil)
}

func (c *ElementoClient) CreateStorageResource(storageResource *StorageResource) (CreateResponse, error) {
    uuid, err := c.getClientUUID()
    if err != nil {
        return nil, err
    }
    storageResource.CreatorUUID = *uuid

    return c.createResource(storageResource, ":7772/api/v1.0/create", "storage", nil)
}

func (c *ElementoClient) DeleteStorageResource(volume_uuid string) (int, error) {
	client_uuid, err := c.getClientUUID()
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

    return c.deleteResource(delRequest, "POST", ":7772/api/v1.0/destroy", "storage", nil)
}

func (c *ElementoClient) ListStorageResources() ([]StorageResourceResponse, error) {
	client_uuid, err := c.getClientUUID()
	if err != nil {
		return nil, err
	}

	var listRequest = struct {
		ClientUUID string `json:"client_uid"`
	}{
		ClientUUID: *client_uuid,
	}

    return listResources[StorageResourceResponse](c, listRequest, "POST", ":7772/api/v1.0/accessible", "storage", nil)
}

func (c *ElementoClient) GetStorageResource(volume_uuid string) (*StorageResourceResponse, int, error) {
	client_uuid, err := c.getClientUUID()
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

    return getResource[StorageResourceResponse](c, getRequest, "POST", ":7772/api/v1.0/info", "storage", nil)
}

func (c *ElementoClient) CanCreateObjectStorageResource(objectStorageResource ObjectStorageRequest) (*CanCreateResponse, error) {
    return c.canCreate(objectStorageResource, "POST", ":7770/api/v1.0/objectstorage/cancreate", "objectstorage", nil)
}

func (c *ElementoClient) CreateObjectStorageResource(objectStorageResource ObjectStorageRequest) (CreateResponse, error) {
    return c.createResource(objectStorageResource, ":7770/api/v1.0/objectstorage/create", "objectstorage", nil)
}

func (c *ElementoClient) DeleteObjectStorageResource(service_uuid string) (int, error) {
	delRequest := struct {
		UUID string `json:"service_uuid"`
	}{
		UUID: service_uuid,
	}

    return c.deleteResource(delRequest, "DELETE", ":7770/api/v1.0/objectstorage/delete", "objectstorage", nil)
}

func (c *ElementoClient) ListObjectStorageResources() ([]ObjectStorageResponse, error) {
    return listResources[ObjectStorageResponse](c, nil, "GET", ":7770/api/v1.0/objectstorage/running", "objectstorage", nil)
}

func (c *ElementoClient) GetObjectStorageResource(service_uuid string) (*ObjectStorageResponse, int, error) {
    return getResource[ObjectStorageResponse](c, nil, "GET", ":7770/api/v1.0/objectstorage/running/" + service_uuid, "objectstorage", nil)
}

func (c *ElementoClient) GetObjectStorageCredentials(service_uuid string) (*ObjectStorageCredentials, error) {
	urlSuffix := ":7770/api/v1.0/objectstorage/credentials/" + service_uuid

	serializer := func(r io.Reader) (*ObjectStorageCredentials, error) {
		return KVDeserializer[ObjectStorageCredentials](r, "=")
	}

    return getCredentials[ObjectStorageCredentials](c, urlSuffix, "objectstorage", serializer, nil)
}

func (c *ElementoClient) CanCreateDBAASResource(dbaasResource DBAASRequest) (*CanCreateResponse, error) {
    return c.canCreate(dbaasResource, "POST", ":7770/api/v1.0/dbaas/cancreate", "dbaas", nil)
}

func (c *ElementoClient) CreateDBAASResource(dbaasResource DBAASRequest) (CreateResponse, error) {
    return c.createResource(dbaasResource, ":7770/api/v1.0/dbaas/create", "dbaas", nil)
}

func (c *ElementoClient) DeleteDBAASResource(service_uuid string) (int, error) {
	delRequest := struct {
		UUID string `json:"service_uuid"`
	}{
		UUID: service_uuid,
	}

    return c.deleteResource(delRequest, "DELETE", ":7770/api/v1.0/dbaas/delete", "dbaas", nil)
}

func (c *ElementoClient) ListDBAASResources() ([]DBAASResponse, error) {
    return listResources[DBAASResponse](c, nil, "GET", ":7770/api/v1.0/dbaas/running", "dbaas", nil)
}

func (c *ElementoClient) GetDBAASResource(service_uuid string) (*DBAASResponse, int, error) {
    return getResource[DBAASResponse](c, nil, "GET", ":7770/api/v1.0/dbaas/running/" + service_uuid, "dbaas", nil)
}

func (c *ElementoClient) GetDBAASCredentials(service_uuid string) (*DBAASCredentials, error) {
	urlSuffix := ":7770/api/v1.0/dbaas/credentials/" + service_uuid

	serializer := func(r io.Reader) (*DBAASCredentials, error) {
		return KVDeserializer[DBAASCredentials](r, "=")
	}

    return getCredentials[DBAASCredentials](c, urlSuffix, "dbaas", serializer, nil)
}



func (c *ElementoClient) CanCreateKAASResource(kaasResource KAASRequest) (*CanCreateResponse, error) {
    return c.canCreate(kaasResource, "POST", ":7770/api/v1.0/kaas/cancreate", "kaas", nil)
}

func (c *ElementoClient) CreateKAASResource(kaasResource KAASRequest) (CreateResponse, error) {
    return c.createResource(kaasResource, ":7770/api/v1.0/kaas/create", "kaas", nil)
}

func (c *ElementoClient) DeleteKAASResource(service_uuid string) (int, error) {
	delRequest := struct {
		UUID string `json:"service_uuid"`
	}{
		UUID: service_uuid,
	}

    return c.deleteResource(delRequest, "DELETE", ":7770/api/v1.0/kaas/delete", "kaas", nil)
}

func (c *ElementoClient) ListKAASResources() ([]KAASResponse, error) {
    return listResources[KAASResponse](c, nil, "GET", ":7770/api/v1.0/kaas/running", "kaas", nil)
}

func (c *ElementoClient) GetKAASResource(service_uuid string) (*KAASResponse, int, error) {
    return getResource[KAASResponse](c, nil, "GET", ":7770/api/v1.0/kaas/running/" + service_uuid, "kaas", nil)
}

func (c *ElementoClient) GetKAASCredentials(service_uuid string) (*string, error) {
	urlSuffix := ":7770/api/v1.0/kaas/credentials/" + service_uuid

	deserializer := func(r io.Reader) (*string, error) {
		credentials, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

        creds := string(credentials)
        return &creds, nil
    }
    return getCredentials[string](c, urlSuffix, "kaas", deserializer, nil)
}


// func (c *ElementoClient) ListN8NResources() ([]ComputeListItem, error) {
//     ret, err := c.ListComputeResources()
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

func (c *ElementoClient) GetN8NCredentials(service_uuid string) (*N8NCredentials, error) {
	urlSuffix := ":7777/api/v1.0/credentials/" + service_uuid

	serializer := func(r io.Reader) (*N8NCredentials, error) {
		return KVDeserializer[N8NCredentials](r, ": |=")
	}

    return getCredentials[N8NCredentials](c, urlSuffix, "n8n", serializer, nil)
}


// func (c *ElementoClient) ListOpenClawResources() ([]ComputeListItem, error) {
//     ret, err := c.ListComputeResources()
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

func (c *ElementoClient) GetOpenClawCredentials(service_uuid string) (*OpenClawCredentials, error) {
	urlSuffix := ":7777/api/v1.0/credentials/" + service_uuid

	serializer := func(r io.Reader) (*OpenClawCredentials, error) {
		return KVDeserializer[OpenClawCredentials](r, ": |=")
	}

    return getCredentials[OpenClawCredentials](c, urlSuffix, "openclaw", serializer, nil)
}


func (c *ElementoClient) CanCreateFileStorageResource(fileStorageResource FileStorageRequest) (*CanCreateResponse, error) {
    return c.canCreate(fileStorageResource, "POST", ":7770/api/v1.0/filestorage/cancreate", "filestorage", nil)
}

func (c *ElementoClient) CreateFileStorageResource(fileStorageResource FileStorageRequest) (CreateResponse, error) {
    return c.createResource(fileStorageResource, ":7770/api/v1.0/filestorage/create", "filestorage", nil)
}

func (c *ElementoClient) DeleteFileStorageResource(service_uuid string) (int, error) {
    delRequest := struct {
        UUID       string `json:"service_uuid"`
    }{
        UUID: service_uuid,
    }

    return c.deleteResource(delRequest, "DELETE", ":7770/api/v1.0/filestorage/delete", "filestorage", nil)
}

func (c *ElementoClient) ListFileStorageResources() ([]FileStorageResponse, error) {
    return listResources[FileStorageResponse](c, nil, "GET", ":7770/api/v1.0/filestorage/running", "filestorage", nil)
}

func (c *ElementoClient) GetFileStorageResource(service_uuid string) (*FileStorageResponse, int, error) {
    return getResource[FileStorageResponse](c, nil, "GET", ":7770/api/v1.0/filestorage/running/" + service_uuid, "filestorage", nil)
}

func (c *ElementoClient) GetFileStorageCredentials(service_uuid string) (*FileStorageCredentials, error) {
    urlSuffix := ":7770/api/v1.0/filestorage/credentials/" + service_uuid

    serializer := func (r io.Reader) (*FileStorageCredentials, error) {
        return KVDeserializer[FileStorageCredentials](r, "=")
    }

    return getCredentials[FileStorageCredentials](c, urlSuffix, "filestorage", serializer, nil)
}


func (c *ElementoClient) CanCreatePrivateNetworkResource(privateNetworkResource PrivateNetworkRequest) (*CanCreateResponse, error) {
    return c.canCreate(privateNetworkResource, "POST", ":7770/api/v1.0/privatenetwork/cancreate", "privatenetwork", nil)
}

func (c *ElementoClient) CreatePrivateNetworkResource(privateNetworkResource PrivateNetworkRequest) (CreateResponse, error) {
    return c.createResource(privateNetworkResource, ":7770/api/v1.0/privatenetwork/create", "privatenetwork", nil)
}

func (c *ElementoClient) DeletePrivateNetworkResource(service_uuid string) (int, error) {
    delRequest := struct {
        UUID       string `json:"service_uuid"`
    }{
        UUID: service_uuid,
    }

    return c.deleteResource(delRequest, "DELETE", ":7770/api/v1.0/privatenetwork/delete", "privatenetwork", nil)
}

func (c *ElementoClient) ListPrivateNetworkResources() ([]PrivateNetworkResponse, error) {
    return listResources[PrivateNetworkResponse](c, nil, "GET", ":7770/api/v1.0/privatenetwork/running", "privatenetwork", nil)
}

func (c *ElementoClient) GetPrivateNetworkResource(service_uuid string) (*PrivateNetworkResponse, int, error) {
    return getResource[PrivateNetworkResponse](c, nil, "GET", ":7770/api/v1.0/privatenetwork/running/" + service_uuid, "privatenetwork", nil)
}

func (c *ElementoClient) GetPrivateNetworkCredentials(service_uuid string) (*PrivateNetworkCredentials, error) {
    urlSuffix := ":7770/api/v1.0/privatenetwork/credentials/" + service_uuid

    serializer := func (r io.Reader) (*PrivateNetworkCredentials, error) {
        return KVDeserializer[PrivateNetworkCredentials](r, "=")
    }

    return getCredentials[PrivateNetworkCredentials](c, urlSuffix, "privatenetwork", serializer, nil)
}


func (c *ElementoClient) CanCreateResource(resource resourceWithCanCreate) (*CanCreateResponse, error) {
	cfg := resource.canCreateConfig()
    return c.canCreate(resource, cfg.Method, cfg.UrlSuffix, cfg.ResourceName, cfg.Headers)
}

func (c *ElementoClient) CreateResource(resource resourceWithCreate) (CreateResponse, error) {
	cfg := resource.createConfig()
    return c.createResource(resource, cfg.UrlSuffix, cfg.ResourceName, cfg.Headers)
}

func (c *ElementoClient) DeleteResource(service_uuid string, resourceName string) (int, error) {
    delRequest := struct {
        UUID       string `json:"service_uuid"`
    }{
        UUID: service_uuid,
    }

    return c.deleteResource(delRequest, "DELETE", fmt.Sprintf(":7770/api/v1.0/%s/delete", resourceName), resourceName, nil)
}

func ListResources[T any](c *ElementoClient, resourceName string) ([]T, error) {
    return listResources[T](c, nil, "GET", fmt.Sprintf(":7770/api/v1.0/%s/running", resourceName), resourceName, nil)
}

func GetResource[T any](c *ElementoClient, service_uuid string, resourceName string) (*T, int, error) {
    return getResource[T](c, nil, "GET", fmt.Sprintf(":7770/api/v1.0/%s/running/%s", resourceName, service_uuid), resourceName, nil)
}

func (c *ElementoClient) GetRegistryCredentials(service_uuid string) (*RegistryCredentials, error) {
    urlSuffix := ":7770/api/v1.0/registry/credentials/" + service_uuid

    serializer := func (r io.Reader) (*RegistryCredentials, error) {
        return KVDeserializer[RegistryCredentials](r, "=")
    }

    return getCredentials[RegistryCredentials](c, urlSuffix, "registry", serializer, nil)
}

func (c *ElementoClient) GetServerlessCredentials(service_uuid string) (*ServerlessCredentials, error) {
    urlSuffix := ":7770/api/v1.0/serverless/credentials/" + service_uuid

    serializer := func (r io.Reader) (*ServerlessCredentials, error) {
        return KVDeserializer[ServerlessCredentials](r, "=")
    }

    return getCredentials[ServerlessCredentials](c, urlSuffix, "serverless", serializer, nil)
}

func (c *ElementoClient) GetAIServerlessCredentials(service_uuid string) (*AIServerlessCredentials, error) {
    urlSuffix := ":7770/api/v1.0/aiserverless/credentials/" + service_uuid

    serializer := func (r io.Reader) (*AIServerlessCredentials, error) {
        return KVDeserializer[AIServerlessCredentials](r, "=")
    }

    return getCredentials[AIServerlessCredentials](c, urlSuffix, "aiserverless", serializer, nil)
}

func (c *ElementoClient) GetQueueAASCredentials(service_uuid string) (*QueueAASCredentials, error) {
    urlSuffix := ":7770/api/v1.0/queueaas/credentials/" + service_uuid

    serializer := func (r io.Reader) (*QueueAASCredentials, error) {
        return KVDeserializer[QueueAASCredentials](r, "=")
    }

    return getCredentials[QueueAASCredentials](c, urlSuffix, "queueaas", serializer, nil)
}


func (c *ElementoClient) ListSpotVMResources() ([]SpotVMListItem, error) {
	host, err := c.getHost()
	if err != nil {
		return nil, err
	}

	// Legacy compute handling needs special handling as the return payload
	// is in the form of {VMS: [VMItem]} rather than just [VMItem]
    url := "https://" + *host + ":7777/api/v1.0/running"

    authToken, err := c.getAuth()
    if err != nil {
        return nil, err
    }

    target_uuid, err := c.getTargetUUID()
    if err != nil {
    	return nil, err
    }

    headers := map[string]string{
        "Authorization": "Bearer " + *authToken,
        "X-Target-UUID": *target_uuid,
    }

    resp, err := c.request("GET", url, headers, nil)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

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

func (c *ElementoClient) DeleteSpotVMResource(uuid string) (int, error) {
    delRequest := struct {
        UUID string `json:"vm_uid"`
    }{
        UUID: uuid,
    }
    return c.deleteResource(delRequest, "DELETE", ":7777/api/v1.0/unregister", "spotvm", nil)
}
