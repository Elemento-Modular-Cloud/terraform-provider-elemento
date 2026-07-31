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
)

type ElementoClient struct {
	Host     string
	Username string
	Password string
	Portal  string
	httpClient *http.Client
	cachedToken *string
	cachedTokenExpiry int64
	cachedUUID *string
}

func NewElementoClient(host, username, password, portal string) *ElementoClient {
    httpClient := &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
            },
        },
    }

    return &ElementoClient{
		Host:     host,
		Username: username,
		Password: password,
		Portal:   portal,
		httpClient: httpClient,
	}
}

type ComputeResponse struct {
    PaymentURL  string `json:"payment_url"`
    BillingUUID string `json:"billing_uuid"`
}

type CreateResponse struct {
    PaymentURL  string `json:"payment_url"`
    BillingUUID string `json:"billing_uuid"`
}

type CanCreateBilling struct {
    BillingUUID string `json:"billing_uuid"`
    PaymentURL  string `json:"payment_url"`
}

type CanCreateResponse struct {
    CanCreate bool               `json:"cancreate"`
    Billing   []CanCreateBilling `json:"billing"`
    Provider  string             `json:"provider"`
    Reasons   map[string]string  `json:"reasons"`
}

type authRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type authResponse struct {
    Username       string `json:"username"`
    AccessToken    string `json:"access_token"`
    RefreshToken   string `json:"refresh_token"`
    AccessTokenIAT int64  `json:"access_token_iat"`
    AccessTokenExp int64  `json:"access_token_exp"`
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

func (c *ElementoClient) getAuth() (*string, error) {
    if c.cachedToken != nil && c.cachedTokenExpiry > time.Now().Unix() {
        return c.cachedToken, nil
    }

    url := "https://" + c.Portal + "/api/v2/auth/userpass"

    authRequest := authRequest {
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

    defer resp.Body.Close()

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

func (c *ElementoClient) getClientUUID() (*string, error) {
    if c.cachedUUID != nil {
        return c.cachedUUID, nil
    }

    url := "https://" + c.Portal + "/api/v2/account/uuid"

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
        return nil, fmt.Errorf("failed to get client UUID: %s %s", resp.Status, string(body))
    }

    var uuidResponse struct {
        UUID string `json:"id"`
    }
    err = json.NewDecoder(resp.Body).Decode(&uuidResponse)
    if err != nil {
        return nil, err
    }

    c.cachedUUID = &uuidResponse.UUID

    return c.cachedUUID, nil
}

func (c *ElementoClient) canCreate(resource any, method, urlSuffix, resourceName string) (*CanCreateResponse, error) {
    url := "https://" + c.Host + urlSuffix

    authToken, err := c.getAuth()
    if err != nil {
        return nil, err
    }

    headers := map[string]string{
        "Content-Type": "application/json",
        "Authorization": "Bearer " + *authToken,
    }
    resp, err := c.request(method, url, headers, resource)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

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
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("failed to get client UUID: %s %s", resp.Status, string(body))
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

func (c *ElementoClient) createResource(resource any, urlSuffix, resourceName string) (*CreateResponse, error) {
    url := "https://" + c.Host + urlSuffix

    authToken, err := c.getAuth()
    if err != nil {
        return nil, err
    }

    headers := map[string]string{
        "Content-Type": "application/json",
        "Authorization": "Bearer " + *authToken,
    }
    resp, err := c.request("POST", url, headers, resource)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("failed to create %s resource: %s %s", resourceName, resp.Status, string(body))
    }

    var response CreateResponse
    err = json.NewDecoder(resp.Body).Decode(&response)
    if err != nil {
        return nil, err
    }

    return &response, nil
}

func (c *ElementoClient) deleteResource(deleteRequest any, method, urlSuffix, resourceName string) error {
    url := "https://" + c.Host + urlSuffix

    authToken, err := c.getAuth()
    if err != nil {
        return err
    }

    headers := map[string]string{
        "Content-Type": "application/json",
        "Authorization": "Bearer " + *authToken,
    }
    resp, err := c.request(method, url, headers, deleteRequest)
    if err != nil {
        return err
    }

    defer resp.Body.Close()


    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed to delete %s resource: %s %s", resourceName, resp.Status, string(body))
    }

    return nil
}

func listResources[T any](c *ElementoClient, listRequest any, method, urlSuffix, resourceName string) ([]T, error) {
    url := "https://" + c.Host + urlSuffix

    authToken, err := c.getAuth()
    if err != nil {
        return nil, err
    }

    headers := map[string]string{
        "Authorization": "Bearer " + *authToken,
    }
    if listRequest != nil {
        headers["Content-Type"] = "application/json"
    }

    resp, err := c.request(method, url, headers, listRequest)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

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

func getResource[T any](c *ElementoClient, getRequest any, method, urlSuffix, resourceName string) (*T, error) {
    url := "https://" + c.Host + urlSuffix

    authToken, err := c.getAuth()
    if err != nil {
        return nil, err
    }

    headers := map[string]string{
        "Authorization": "Bearer " + *authToken,
    }
    if getRequest != nil {
        headers["Content-Type"] = "application/json"
    }

    resp, err := c.request(method, url, headers, getRequest)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

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

func getCredentials[T any](c *ElementoClient, urlSuffix, resourceName string, deserialize deserializer[T]) (*T, error) {
    url := "https://" + c.Host + urlSuffix

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
        return nil, fmt.Errorf("failed to get %s credentials: %s %s", resourceName, resp.Status, string(body))
    }

    resource, err := deserialize(resp.Body)
    if err != nil {
        return nil, err
    }

    return resource, nil
}


func (c *ElementoClient) CanCreateComputeResource(computeResource ComputeResource) (*CanCreateResponse, error) {
    return c.canCreate(computeResource, "GET", ":7777/api/v1.0/canallocate", "compute")
}

func (c *ElementoClient) CreateComputeResource(computeResource ComputeResource) (*CreateResponse, error) {
    return c.createResource(computeResource, ":7777/api/v1.0/register", "compute")
}

func (c *ElementoClient) ListComputeResources() ([]ComputeListItem, error) {
    url := "https://" + c.Host + ":7777/api/v1.0/running"

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
        return nil, fmt.Errorf("failed to list compute resources: %s %s", resp.Status, string(body))
    }

    var computeResources struct {
        VMS []ComputeListItem `json:"vms"`
    }

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
    return c.deleteResource(delRequest, "DELETE", ":7777/api/v1.0/unregister", "compute")
}


func (c *ElementoClient) CanCreateStorageResource(req CanCreateStorageRequest) (*CanCreateResponse, error) {
    return c.canCreate(req, "GET", ":7772/api/v1.0/cancreate", "storage")
}

func (c *ElementoClient) CreateStorageResource(storageResource StorageResource) (*CreateResponse, error) {
    uuid, err := c.getClientUUID()
    if err != nil {
        return nil, err
    }
    storageResource.CreatorUUID = *uuid

    return c.createResource(storageResource, ":7772/api/v1.0/create", "storage")
}

func (c *ElementoClient) DeleteStorageResource(client_uuid, volume_uuid string) error {
    delRequest := struct {
        UUID       string `json:"volume_id"`
        ClientUUID string `json:"client_id"`
    }{
        UUID: volume_uuid,
        ClientUUID: client_uuid,
    }

    return c.deleteResource(delRequest, "POST", ":7772/api/v1.0/destroy", "storage")
}

func (c *ElementoClient) ListStorageResources(client_uuid string) ([]StorageResourceResponse, error) {
    var listRequest = struct {
        ClientUUID string `json:"client_uid"`
    }{
        ClientUUID: client_uuid,
    }

    return listResources[StorageResourceResponse](c, listRequest, "POST", ":7772/api/v1.0/accessible", "storage")
}

func (c *ElementoClient) GetStorageResource(client_uuid, volume_uuid string) (*StorageResourceResponse, error) {
    var getRequest = struct {
        ClientUUID string `json:"client_uid"`
        VolumeUUID string `json:"volume_id"`
    }{
        ClientUUID: client_uuid,
        VolumeUUID: volume_uuid,
    }

    return getResource[StorageResourceResponse](c, getRequest, "POST", ":7772/api/v1.0/info", "storage")
}


func (c *ElementoClient) CanCreateObjectStorageResource(objectStorageResource ObjectStorageRequest) (*CanCreateResponse, error) {
    return c.canCreate(objectStorageResource, "POST", ":7770/api/v1.0/objectstorage/cancreate", "objectstorage")
}

func (c *ElementoClient) CreateObjectStorageResource(objectStorageResource ObjectStorageRequest) (*CreateResponse, error) {
    return c.createResource(objectStorageResource, ":7770/api/v1.0/objectstorage/create", "objectstorage")
}

func (c *ElementoClient) DeleteObjectStorageResource(service_uuid string) error {
    delRequest := struct {
        UUID       string `json:"service_uuid"`
    }{
        UUID: service_uuid,
    }

    return c.deleteResource(delRequest, "DELETE", ":7770/api/v1.0/objectstorage/delete", "objectstorage")
}

func (c *ElementoClient) ListObjectStorageResources() ([]ObjectStorageResponse, error) {
    return listResources[ObjectStorageResponse](c, nil, "GET", ":7770/api/v1.0/objectstorage/running", "objectstorage")
}

func (c *ElementoClient) GetObjectStorageResource(service_uuid string) (*ObjectStorageResponse, error) {
    return getResource[ObjectStorageResponse](c, nil, "GET", ":7770/api/v1.0/objectstorage/running/" + service_uuid, "objectstorage")
}

func (c *ElementoClient) GetObjectStorageCredentials(service_uuid string) (*ObjectStorageCredentials, error) {
    urlSuffix := ":7770/api/v1.0/objectstorage/credentials/" + service_uuid

    serializer := func (r io.Reader) (*ObjectStorageCredentials, error) {
        return KVDeserializer[ObjectStorageCredentials](r, "=")
    }

    return getCredentials[ObjectStorageCredentials](c, urlSuffix, "objectstorage", serializer)
}


func (c *ElementoClient) CanCreateDBAASResource(dbaasResource DBAASRequest) (*CanCreateResponse, error) {
    return c.canCreate(dbaasResource, "POST", ":7770/api/v1.0/dbaas/cancreate", "dbaas")
}

func (c *ElementoClient) CreateDBAASResource(dbaasResource DBAASRequest) (*CreateResponse, error) {
    return c.createResource(dbaasResource, ":7770/api/v1.0/dbaas/create", "dbaas")
}

func (c *ElementoClient) DeleteDBAASResource(service_uuid string) error {
    delRequest := struct {
        UUID       string `json:"service_uuid"`
    }{
        UUID: service_uuid,
    }

    return c.deleteResource(delRequest, "DELETE", ":7770/api/v1.0/dbaas/delete", "dbaas")
}

func (c *ElementoClient) ListDBAASResources() ([]DBAASResponse, error) {
    return listResources[DBAASResponse](c, nil, "GET", ":7770/api/v1.0/dbaas/running", "dbaas")
}

func (c *ElementoClient) GetDBAASResource(service_uuid string) (*DBAASResponse, error) {
    return getResource[DBAASResponse](c, nil, "GET", ":7770/api/v1.0/dbaas/running/" + service_uuid, "dbaas")
}

func (c *ElementoClient) GetDBAASCredentials(service_uuid string) (*DBAASCredentials, error) {
    urlSuffix := ":7770/api/v1.0/dbaas/credentials/" + service_uuid

    serializer := func (r io.Reader) (*DBAASCredentials, error) {
        return KVDeserializer[DBAASCredentials](r, "=")
    }

    return getCredentials[DBAASCredentials](c, urlSuffix, "dbaas", serializer)
}



func (c *ElementoClient) CanCreateKAASResource(kaasResource KAASRequest) (*CanCreateResponse, error) {
    return c.canCreate(kaasResource, "POST", ":7770/api/v1.0/kaas/cancreate", "kaas")
}

func (c *ElementoClient) CreateKAASResource(kaasResource KAASRequest) (*CreateResponse, error) {
    return c.createResource(kaasResource, ":7770/api/v1.0/kaas/create", "kaas")
}

func (c *ElementoClient) DeleteKAASResource(service_uuid string) error {
    delRequest := struct {
        UUID       string `json:"service_uuid"`
    }{
        UUID: service_uuid,
    }

    return c.deleteResource(delRequest, "DELETE", ":7770/api/v1.0/kaas/delete", "kaas")
}

func (c *ElementoClient) ListKAASResources() ([]KAASResponse, error) {
    return listResources[KAASResponse](c, nil, "GET", ":7770/api/v1.0/kaas/running", "kaas")
}

func (c *ElementoClient) GetKAASResource(service_uuid string) (*KAASResponse, error) {
    return getResource[KAASResponse](c, nil, "GET", ":7770/api/v1.0/kaas/running/" + service_uuid, "kaas")
}

func (c *ElementoClient) GetKAASCredentials(service_uuid string) (*string, error) {
    deserializer := func (r io.Reader) (*string, error) {
        credentials, err := io.ReadAll(r)
        if err != nil {
            return nil, err
        }

        creds := string(credentials)
        return &creds, nil
    }
    return getCredentials[string](c, "kaas", service_uuid, deserializer)
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

    serializer := func (r io.Reader) (*N8NCredentials, error) {
        return KVDeserializer[N8NCredentials](r, ": ")
    }

    return getCredentials[N8NCredentials](c, urlSuffix, "n8n", serializer)
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

    serializer := func (r io.Reader) (*OpenClawCredentials, error) {
        return KVDeserializer[OpenClawCredentials](r, ": ")
    }

    return getCredentials[OpenClawCredentials](c, urlSuffix, "openclaw", serializer)
}
