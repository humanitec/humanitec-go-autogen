package humanitec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/humanitec/humanitec-go-autogen/client"
)

const (
	DefaultAPIHost = "https://api.humanitec.io/"
	SDK            = "humanitec-go-autogen"
	SDKVersion     = "latest"
)

var (
	SDKHeader       = fmt.Sprintf("%s/%s", SDK, SDKVersion)
	ErrMissingToken = errors.New("token is required")
)

type Config struct {
	// Token used for API requests
	Token string

	// URL is the base URL for the API requests (optional)
	URL string

	// Callbacks for logging requests and responses (optional)
	RequestLogger  func(r *RequestDetails)
	ResponseLogger func(r *ResponseDetails)

	// Use a custom HTTP client (optional)
	Client client.HttpRequestDoer

	// Skip initial token check (optional)
	SkipInitialTokenCheck bool

	// Internal usage tracking (optional)
	InternalApp string
}

type RequestDetails struct {
	Context context.Context
	Method  string
	URL     *url.URL
	Body    []byte
}

type ResponseDetails struct {
	Context    context.Context
	StatusCode int
	Body       []byte
}

type Client struct {
	client.ClientWithResponses
	client *client.Client
}

func (c *Client) Client() *client.Client {
	return c.client
}

func NewClient(config *Config) (*Client, error) {
	if config.Token == "" && !config.SkipInitialTokenCheck {
		return nil, ErrMissingToken
	}

	if config.URL == "" {
		config.URL = DefaultAPIHost
	}

	var doer client.HttpRequestDoer
	if config.Client == nil {
		doer = &http.Client{}
	} else {
		doer = config.Client
	}

	baseClient, err := client.NewClient(config.URL, func(c *client.Client) error {
		c.Client = &DoWithLog{doer, config.RequestLogger, config.ResponseLogger}
		c.RequestEditors = append(c.RequestEditors, func(_ context.Context, req *http.Request) error {
			if config.Token == "" {
				return ErrMissingToken
			}

			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))
			req.Header.Add("Humanitec-User-Agent", humanitecUserAgent(config.InternalApp, SDKHeader))
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Client{client.ClientWithResponses{ClientInterface: baseClient}, baseClient}, nil
}

func humanitecUserAgent(app, sdk string) string {
	parts := []string{}

	if app != "" {
		parts = append(parts, fmt.Sprintf("app %s", app))
	}
	if sdk != "" {
		parts = append(parts, fmt.Sprintf("sdk %s", sdk))
	}

	return strings.Join(parts, "; ")
}

func copyBody(body io.ReadCloser) (io.ReadCloser, []byte, error) {
	if body == nil {
		return nil, nil, nil
	}

	var buf bytes.Buffer
	tee := io.TeeReader(body, &buf)
	bodyBytes, err := io.ReadAll(tee)
	if err != nil {
		return nil, nil, err
	}

	return io.NopCloser(bytes.NewReader(buf.Bytes())), bodyBytes, nil
}

func copyReqBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, nil
	}

	body, bodyBytes, err := copyBody(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = body

	return bodyBytes, nil
}

func copyResBody(res *http.Response) ([]byte, error) {
	if res.Body == nil {
		return nil, nil
	}

	body, bodyBytes, err := copyBody(res.Body)
	if err != nil {
		return nil, err
	}
	res.Body = body

	return bodyBytes, nil
}

type DoWithLog struct {
	client         client.HttpRequestDoer
	requestLogger  func(r *RequestDetails)
	responseLogger func(r *ResponseDetails)
}

func (d *DoWithLog) Do(req *http.Request) (*http.Response, error) {
	if d.requestLogger != nil {
		reqBody, err := copyReqBody(req)
		if err != nil {
			return nil, err
		}

		d.requestLogger(&RequestDetails{
			Context: req.Context(),
			Method:  req.Method,
			URL:     req.URL,
			Body:    reqBody,
		})
	}

	res, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}

	if d.responseLogger != nil {
		resBody, err := copyResBody(res)
		if err != nil {
			return nil, err
		}

		d.responseLogger(&ResponseDetails{
			Context:    req.Context(),
			StatusCode: res.StatusCode,
			Body:       resBody,
		})
	}

	return res, nil
}
