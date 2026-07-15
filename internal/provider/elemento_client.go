package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
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

type ComputeResource struct {
    Requirements     ComputeRequirements  `json:"req"`
    VMName           string               `json:"vm_name"`
    Region           string               `json:"region"`
    CloudInit        CloudInitConfig      `json:"cloudinit"`
    AdditionalInfo   AdditionalInfoConfig `json:"additional_info"`
    BillingFrequency string               `json:"billing_frequency"`
    Volumes          []StorageVolume       `json:"volumes"`
}

type ComputeRequirements struct {
    CPU    ComputeCPU    `json:"cpu"`
    Memory ComputeMemory `json:"mem"`
    PCI    ComputePCI    `json:"pci"`
    Misc   ComputeMisc   `json:"misc"`
}

type ComputeCPU struct {
    SMTEnabled       bool     `json:"smtEnabled"`
    MinFrequency     float64  `json:"min_frequency"`
    Arch             string   `json:"arch"`
    Flags            []string `json:"flags"`
    Slots            int      `json:"slots"`
    AllowSMT         bool     `json:"allowSMT"`
    FullPhysical     bool     `json:"fullPhysical"`
    MaxOverprovision int      `json:"maxOverprovision"`
}

type ComputeMemory struct {
    Capacity int    `json:"capacity"`
    RequireECC bool `json:"requireECC"`
}

type ComputePCI struct {
    Devices map[string]int `json:"devices"`
}

type ComputeMisc struct {
    OSFamily string  `json:"os_family"`
    OSFlavour string `json:"os_flavour"`
}

type CloudInitConfig struct {
    Template  string            `json:"template"`
    Variables map[string]string `json:"variables"`
}

type AdditionalInfoConfig struct {
    Service string `json:"service"`
}

type StorageVolume struct {
    Name      string `json:"name"`
    Size      int    `json:"size"`
    Private   bool   `json:"private"`
    Shareable bool   `json:"shareable"`
    Bootable  bool   `json:"bootable"`
    Readonly  bool   `json:"readonly"`
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

func (c *ElementoClient) getAuth(ctx context.Context) (*string, error) {
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

    tflog.Debug(ctx, "Getting auth from " + url)
    resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    tflog.Debug(ctx, "Got auth from " + url)

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

func (c *ElementoClient) CreateComputeResource(ctx context.Context, computeResource ComputeResource) (*ComputeResponse, error) {
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

    tflog.Debug(ctx, "Getting auth")
    authToken, err := c.getAuth(ctx)
    if err != nil {
        return nil, err
    }
    req.Header.Set("Authorization", "Bearer " + *authToken)
    tflog.Debug(ctx, "Got auth")

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
