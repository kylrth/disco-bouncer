// Package client provides methods for accessing the discord bouncer REST API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/net/publicsuffix"
)

// Client implements methods to access the disco-bouncer API.
type Client struct {
	baseURL string
	client  *http.Client

	Admin AdminService
	Users UsersService
}

// NewClient creates a new client that sends requests to the given URL.
func NewClient(baseURL string) (*Client, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	c := Client{
		baseURL: baseURL,
		client: &http.Client{
			Jar:     jar,
			Timeout: 5 * time.Second,
		},
	}
	c.Admin = AdminService{&c}
	c.Users = UsersService{&c}

	return &c, nil
}

// ErrNotLoggedIn is returned when a 401: Unauthorized is returned by the server.
var ErrNotLoggedIn = errors.New("not logged in")

func handleNotOK(resp *http.Response) error {
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrNotLoggedIn
	}

	errMsg, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read error message body: %w", err)
	}

	return errors.New(string(errMsg))
}

// joinURL adds the endpoint path to the end of the baseURL. endpoint may contain query parameters.
func joinURL(baseURL, endpoint string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	endp, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}

	base.Path, err = url.JoinPath(base.Path, endp.Path)
	if err != nil {
		return "", err
	}

	base.RawQuery = endp.RawQuery

	return base.String(), nil
}

func (c *Client) get(ctx context.Context, p string) (*http.Response, error) {
	p, err := joinURL(c.baseURL, p)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, handleNotOK(resp)
	}

	return resp, nil
}

func unmarshalBody(resp *http.Response, v interface{}) error {
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (c *Client) getJSON(ctx context.Context, p string, v interface{}) error {
	resp, err := c.get(ctx, p)
	if err != nil {
		return err
	}

	return unmarshalBody(resp, v)
}

func (c *Client) post(
	ctx context.Context, p, contentType string, body io.Reader,
) (*http.Response, error) {
	p, err := joinURL(c.baseURL, p)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, handleNotOK(resp)
	}

	return resp, nil
}

func (c *Client) postJSON(ctx context.Context, p string, data interface{}) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return c.post(ctx, p, "application/json", bytes.NewBuffer(body))
}

func (c *Client) postJSONrecvJSON(ctx context.Context, p string, data, v interface{}) error {
	resp, err := c.postJSON(ctx, p, data)
	if err != nil {
		return err
	}

	return unmarshalBody(resp, v)
}

func (c *Client) putJSONrecvJSON(ctx context.Context, p string, data, v interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	p, err = joinURL(c.baseURL, p)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, p, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return handleNotOK(resp)
	}

	return unmarshalBody(resp, v)
}

func (c *Client) delete(ctx context.Context, p string) error {
	p, err := joinURL(c.baseURL, p)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, p, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return handleNotOK(resp)
	}

	return nil
}
