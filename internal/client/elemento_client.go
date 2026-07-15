package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

type authRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type authResponse struct {
    Username string `json:"username"`
    AccessToken string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    AccessTokenIAT int64 `json:"access_token_iat"`
    AccessTokenExp int64 `json:"access_token_exp"`
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

    body, err := json.Marshal(authRequest)
    if err != nil {
        return nil, err
    }

    resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
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

func (c *ElementoClient) CreateComputeResource(computeResource ComputeResource) (*ComputeResponse, error) {
    url := "https://" + c.Host + ":7777/api/v1.0/register"

    body, err := json.Marshal(computeResource)
    if err != nil {
        return nil, err
    }

    req, err := http.NewRequest("POST", url, bytes.NewReader(body))
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", "application/json")

    authToken, err := c.getAuth()
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer " + *authToken)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        body, _ := io.ReadAll(resp.Body)

        return nil, fmt.Errorf("failed to create compute resource: %s %s", resp.Status, string(body))
    }

    var response ComputeResponse
    err = json.NewDecoder(resp.Body).Decode(&response)
    if err != nil {
        return nil, err
    }

    return &response, nil
}

func (c *ElementoClient) ListComputeResources(ctx context.Context) ([]ComputeListItem, error) {
    url := "https://" + c.Host + ":7777/api/v1.0/running"

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }

    authToken, err := c.getAuth()
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer " + *authToken)

    resp, err := c.httpClient.Do(req)
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
