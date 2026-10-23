package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
	// "github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/golang-jwt/jwt/v5"
)

type ElementoClient struct {
	Target            string
	Username          string
	Password          string
	Portal            string
	httpClient        *http.Client
	cachedToken       *string
	cachedTokenExpiry int64
	cachedUUID        *string
	targetData        *TargetGrant
}

func NewElementoClient(target, username, password, portal string) *ElementoClient {
    httpClient := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
            },
        },
    }

    return &ElementoClient{
		Target:     target,
		Username: username,
		Password: password,
		Portal:   portal,
		httpClient: httpClient,
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
	if c.cachedToken != nil && c.cachedTokenExpiry > time.Now().Unix() {
		return c.cachedToken, nil
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

	c.cachedToken = &authResponse.AccessToken
	c.cachedTokenExpiry = authResponse.AccessTokenExp
	return c.cachedToken, nil
}

// ClientUUID returns the authenticated account UUID from the Portal.
func (c *ElementoClient) ClientUUID() (*string, error) {
    return c.getClientUUID()
}

func (c *ElementoClient) getClientUUID() (*string, error) {
	if c.cachedUUID != nil {
		return c.cachedUUID, nil
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

    c.cachedUUID = &clientUUID

	return c.cachedUUID, nil
}

func (c *ElementoClient) getTargetData() (*TargetGrant, error) {
	if c.targetData != nil {
		return c.targetData, nil
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

    if resp.StatusCode == 202 {
        var response PublicCreateResponse
        err = json.NewDecoder(resp.Body).Decode(&response)
        if err != nil {
            return nil, err
        }
        return response, nil
    } else if resp.StatusCode == 200 {
        var response PrivateCreateResponse
        err = json.NewDecoder(resp.Body).Decode(&response)
        if err != nil {
            return nil, err
        }
        return response, nil
    }

    return nil, fmt.Errorf("unexpected status code %d when creating %s resource", resp.StatusCode, resourceName)
}

func (c *ElementoClient) deleteResource(deleteRequest any, method, urlSuffix, resourceName string, headers map[string]string) error {
	host, err := c.getHost()
	if err != nil {
		return err
	}
    url := "https://" + *host + urlSuffix

	authToken, err := c.getAuth()
	if err != nil {
		return err
	}

    target_uuid, err := c.getTargetUUID()
    if err != nil {
    	return err
    }

    if headers == nil {
    	headers = map[string]string{}
    }
    headers["Content-Type"] = "application/json"
    headers["Authorization"] = "Bearer " + *authToken
    headers["X-Target-UUID"] = *target_uuid

    resp, err := c.request(method, url, headers, deleteRequest)
    if err != nil {
        return err
    }

    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete %s resource: %s %s", resourceName, resp.Status, string(body))
	}

	return nil
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

func getResource[T any](c *ElementoClient, getRequest any, method, urlSuffix, resourceName string, headers map[string]string) (*T, error) {
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

    if getRequest != nil {
        headers["Content-Type"] = "application/json"
    }

	resp, err := c.request(method, url, headers, getRequest)
	if err != nil {
		return nil, err
	}

    defer closeResponseBody(resp)

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get %s resource: %s %s", resourceName, resp.Status, string(body))
	}

	var resource T

	err = json.NewDecoder(resp.Body).Decode(&resource)
	if err != nil {
		return nil, err
	}

	return &resource, nil
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
		resource = computeResource.toSpotVM()
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
		resource = computeResource.toSpotVM()
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

func (c *ElementoClient) DeleteComputeResource(uuid string) error {
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

func (c *ElementoClient) DeleteStorageResource(client_uuid, volume_uuid string) error {
	delRequest := struct {
		UUID       string `json:"volume_id"`
		ClientUUID string `json:"client_id"`
	}{
		UUID:       volume_uuid,
		ClientUUID: client_uuid,
	}

    return c.deleteResource(delRequest, "POST", ":7772/api/v1.0/destroy", "storage", nil)
}

func (c *ElementoClient) ListStorageResources(client_uuid string) ([]StorageResourceResponse, error) {
	var listRequest = struct {
		ClientUUID string `json:"client_uid"`
	}{
		ClientUUID: client_uuid,
	}

    return listResources[StorageResourceResponse](c, listRequest, "POST", ":7772/api/v1.0/accessible", "storage", nil)
}

func (c *ElementoClient) GetStorageResource(client_uuid, volume_uuid string) (*StorageResourceResponse, error) {
	var getRequest = struct {
		ClientUUID string `json:"client_uid"`
		VolumeUUID string `json:"volume_id"`
	}{
		ClientUUID: client_uuid,
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

func (c *ElementoClient) DeleteObjectStorageResource(service_uuid string) error {
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

func (c *ElementoClient) GetObjectStorageResource(service_uuid string) (*ObjectStorageResponse, error) {
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

func (c *ElementoClient) DeleteDBAASResource(service_uuid string) error {
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

func (c *ElementoClient) GetDBAASResource(service_uuid string) (*DBAASResponse, error) {
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

func (c *ElementoClient) DeleteKAASResource(service_uuid string) error {
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

func (c *ElementoClient) GetKAASResource(service_uuid string) (*KAASResponse, error) {
    return getResource[KAASResponse](c, nil, "GET", ":7770/api/v1.0/kaas/running/" + service_uuid, "kaas", nil)
}

func (c *ElementoClient) GetKAASCredentials(service_uuid string) (*string, error) {
	deserializer := func(r io.Reader) (*string, error) {
		credentials, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

        creds := string(credentials)
        return &creds, nil
    }
    return getCredentials[string](c, "kaas", service_uuid, deserializer, nil)
}

func (c *ElementoClient) ListN8NResources() ([]ComputeListItem, error) {
	ret, err := c.ListComputeResources()
	if err != nil {
		return nil, err
	}

	var resources []ComputeListItem
	for _, item := range ret {
		if slices.Contains(strings.Split(item.Tags, ":"), "n8n") {
			resources = append(resources, item)
		}
	}

	return resources, nil
}

func (c *ElementoClient) GetN8NCredentials(service_uuid string) (*N8NCredentials, error) {
	urlSuffix := ":7777/api/v1.0/credentials/" + service_uuid

	serializer := func(r io.Reader) (*N8NCredentials, error) {
		return KVDeserializer[N8NCredentials](r, ": ")
	}

    return getCredentials[N8NCredentials](c, urlSuffix, "n8n", serializer, nil)
}

func (c *ElementoClient) ListOpenClawResources() ([]ComputeListItem, error) {
	ret, err := c.ListComputeResources()
	if err != nil {
		return nil, err
	}

	var resources []ComputeListItem
	for _, item := range ret {
		if slices.Contains(strings.Split(item.Tags, ":"), "openclaw") {
			resources = append(resources, item)
		}
	}

	return resources, nil
}

func (c *ElementoClient) GetOpenClawCredentials(service_uuid string) (*OpenClawCredentials, error) {
	urlSuffix := ":7777/api/v1.0/credentials/" + service_uuid

	serializer := func(r io.Reader) (*OpenClawCredentials, error) {
		return KVDeserializer[OpenClawCredentials](r, ": ")
	}

    return getCredentials[OpenClawCredentials](c, urlSuffix, "openclaw", serializer, nil)
}
