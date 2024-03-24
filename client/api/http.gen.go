// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.1.0 DO NOT EDIT.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// Devices request
	Devices(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// Folders request
	Folders(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// Completion request
	Completion(ctx context.Context, params *CompletionParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// Health request
	Health(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeviceStats request
	DeviceStats(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// Connections request
	Connections(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) Devices(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDevicesRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Folders(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewFoldersRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Completion(ctx context.Context, params *CompletionParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCompletionRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Health(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewHealthRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeviceStats(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeviceStatsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) Connections(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewConnectionsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewDevicesRequest generates requests for Devices
func NewDevicesRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/rest/config/devices")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewFoldersRequest generates requests for Folders
func NewFoldersRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/rest/config/folders")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCompletionRequest generates requests for Completion
func NewCompletionRequest(server string, params *CompletionParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/rest/db/completion")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryValues := queryURL.Query()

		queryValues.Add("folder", params.Folder)

		queryValues.Add("device", params.Device)

		queryURL.RawQuery = queryValues.Encode()
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewHealthRequest generates requests for Health
func NewHealthRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/rest/noauth/health")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewDeviceStatsRequest generates requests for DeviceStats
func NewDeviceStatsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/rest/stats/device")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewConnectionsRequest generates requests for Connections
func NewConnectionsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/rest/system/connections")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// DevicesWithResponse request
	DevicesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*DevicesResponse, error)

	// FoldersWithResponse request
	FoldersWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*FoldersResponse, error)

	// CompletionWithResponse request
	CompletionWithResponse(ctx context.Context, params *CompletionParams, reqEditors ...RequestEditorFn) (*CompletionResponse, error)

	// HealthWithResponse request
	HealthWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*HealthResponse, error)

	// DeviceStatsWithResponse request
	DeviceStatsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*DeviceStatsResponse, error)

	// ConnectionsWithResponse request
	ConnectionsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ConnectionsResponse, error)
}

type DevicesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]DeviceConfiguration
	JSONDefault  *Error
}

// Status returns HTTPResponse.Status
func (r DevicesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DevicesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type FoldersResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]FolderConfiguration
	JSONDefault  *Error
}

// Status returns HTTPResponse.Status
func (r FoldersResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r FoldersResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CompletionResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *FolderCompletion
	JSONDefault  *Error
}

// Status returns HTTPResponse.Status
func (r CompletionResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CompletionResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type HealthResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *HealthStatus
	JSONDefault  *Error
}

// Status returns HTTPResponse.Status
func (r HealthResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r HealthResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeviceStatsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *map[string]DeviceStatistics
	JSONDefault  *Error
}

// Status returns HTTPResponse.Status
func (r DeviceStatsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeviceStatsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ConnectionsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Connections
	JSONDefault  *Error
}

// Status returns HTTPResponse.Status
func (r ConnectionsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ConnectionsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// DevicesWithResponse request returning *DevicesResponse
func (c *ClientWithResponses) DevicesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*DevicesResponse, error) {
	rsp, err := c.Devices(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDevicesResponse(rsp)
}

// FoldersWithResponse request returning *FoldersResponse
func (c *ClientWithResponses) FoldersWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*FoldersResponse, error) {
	rsp, err := c.Folders(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseFoldersResponse(rsp)
}

// CompletionWithResponse request returning *CompletionResponse
func (c *ClientWithResponses) CompletionWithResponse(ctx context.Context, params *CompletionParams, reqEditors ...RequestEditorFn) (*CompletionResponse, error) {
	rsp, err := c.Completion(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCompletionResponse(rsp)
}

// HealthWithResponse request returning *HealthResponse
func (c *ClientWithResponses) HealthWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*HealthResponse, error) {
	rsp, err := c.Health(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseHealthResponse(rsp)
}

// DeviceStatsWithResponse request returning *DeviceStatsResponse
func (c *ClientWithResponses) DeviceStatsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*DeviceStatsResponse, error) {
	rsp, err := c.DeviceStats(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeviceStatsResponse(rsp)
}

// ConnectionsWithResponse request returning *ConnectionsResponse
func (c *ClientWithResponses) ConnectionsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ConnectionsResponse, error) {
	rsp, err := c.Connections(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseConnectionsResponse(rsp)
}

// ParseDevicesResponse parses an HTTP response from a DevicesWithResponse call
func ParseDevicesResponse(rsp *http.Response) (*DevicesResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DevicesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []DeviceConfiguration
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseFoldersResponse parses an HTTP response from a FoldersWithResponse call
func ParseFoldersResponse(rsp *http.Response) (*FoldersResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &FoldersResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []FolderConfiguration
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseCompletionResponse parses an HTTP response from a CompletionWithResponse call
func ParseCompletionResponse(rsp *http.Response) (*CompletionResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CompletionResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest FolderCompletion
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseHealthResponse parses an HTTP response from a HealthWithResponse call
func ParseHealthResponse(rsp *http.Response) (*HealthResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &HealthResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest HealthStatus
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseDeviceStatsResponse parses an HTTP response from a DeviceStatsWithResponse call
func ParseDeviceStatsResponse(rsp *http.Response) (*DeviceStatsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeviceStatsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest map[string]DeviceStatistics
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}

// ParseConnectionsResponse parses an HTTP response from a ConnectionsWithResponse call
func ParseConnectionsResponse(rsp *http.Response) (*ConnectionsResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ConnectionsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Connections
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && true:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSONDefault = &dest

	}

	return response, nil
}
