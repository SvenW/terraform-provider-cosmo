package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/wundergraph/cosmo/connect-go/gen/proto/wg/cosmo/platform/v1/platformv1connect"
	"github.com/wundergraph/cosmo/terraform-provider-cosmo/internal/utils"
)

type PlatformClient struct {
	Client      platformv1connect.PlatformServiceClient
	cosmoApiKey string
}

func NewClient(apiKey, apiUrl string) (*PlatformClient, error) {
	cosmoApiKey, ok := os.LookupEnv(utils.EnvCosmoApiKey)
	if !ok && apiKey == "" {
		return nil, fmt.Errorf("COSMO_API_KEY environment variable not set and no apiKey provided in provider")
	} 

	if ok && apiKey != "" {
		cosmoApiKey = apiKey
	}

	cosmoApiUrl, ok := os.LookupEnv(utils.EnvCosmoApiUrl)
	if !ok && apiUrl == "" {
		cosmoApiUrl = "https://cosmo-cp.wundergraph.com"
	}

	if ok && apiUrl != "" {
		cosmoApiUrl = apiUrl
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}

	httpClient := &http.Client{
		Transport: &transportWithAuth{
			Transport: transport,
			ApiKey:    cosmoApiKey,
		},
	}

	client := platformv1connect.NewPlatformServiceClient(httpClient, cosmoApiUrl)

	return &PlatformClient{
		Client:      client,
		cosmoApiKey: cosmoApiKey,
	}, nil
}

type transportWithAuth struct {
	Transport http.RoundTripper
	ApiKey    string
}

func (t *transportWithAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.ApiKey)
	return t.Transport.RoundTrip(req)
}
