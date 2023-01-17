package humanitec

import (
	"bytes"
	"context"
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
	SDKHeader = fmt.Sprintf("%s/%s", SDK, SDKVersion)
)

type Config struct {
	Token string
	URL   string

	InternalApp    string
	RequestLogger  func(r *RequestDetails)
	ResponseLogger func(r *ResponseDetails)
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

type Client = client.ClientWithResponses

func NewClient(config *Config) (*Client, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("empty token")
	}

	if config.URL == "" {
		config.URL = DefaultAPIHost
	}

	client, err := client.NewClientWithResponses(config.URL, func(c *client.Client) error {
		c.Client = &DoWithLog{&http.Client{}, config.RequestLogger, config.ResponseLogger}
		c.RequestEditors = append(c.RequestEditors, func(_ context.Context, req *http.Request) error {
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", config.Token))
			req.Header.Add("Humanitec-User-Agent", humanitecUserAgent(config.InternalApp, SDKHeader))
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return client, nil
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
