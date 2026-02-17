package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient struct {
	client  *http.Client
	baseURL string
}

func NewHTTPClient(client *http.Client) *HTTPClient {
	return &HTTPClient{client: client}
}

func Post[ReqBody, ResBody any](client *HTTPClient, path string, reqBody ReqBody) (ResBody, error) {
	var zero ResBody

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return zero, fmt.Errorf("failed to construct request body: %w", err)
	}

	header := http.Header{
		"content-type": []string{"application/json"},
	}

	req, err := http.NewRequest(http.MethodPost, client.baseURL+path, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return zero, fmt.Errorf("failed to construct request: %w", err)
	}
	req.Header = header

	res, err := client.client.Do(req)
	if err != nil {
		return zero, fmt.Errorf("failed to request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("failed to call server, status code: %d, message: ", res.StatusCode)
	}

	resBodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return zero, fmt.Errorf("failed to read response body: %w", err)
	}

	var resBody ResBody
	err = json.Unmarshal(resBodyBytes, &resBody)
	if err != nil {
		return zero, fmt.Errorf("failed to construct response body: %w", err)
	}

	return resBody, nil
}
