package authclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	http    *http.Client
	baseURL string
	prefix  string
	token   string
}

type Options func(*Client)

func WithBaseURL(url string) Options {
	return func(c *Client) {
		c.baseURL = url
	}
}

func WithPrefix(prefix string) Options {
	return func(c *Client) {
		c.prefix = prefix
	}
}

func WithToken(token string) Options {
	return func(c *Client) {
		c.token = token
	}
}

func New(opts ...Options) *Client {
	c := &Client{
		http: &http.Client{
			Timeout: time.Second * 30,
		},
		baseURL: "http://localhost:8090",
		prefix:  "/api",
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) Token() string {
	return c.token
}

// --- Auth types ---

type AuthResponse struct {
	Token  string `json:"token"`
	Record User   `json:"record"`
}

type User struct {
	ID             string `json:"id"`
	Email          string `json:"email"`
	Verified       bool   `json:"verified"`
	CollectionID   string `json:"collectionId"`
	CollectionName string `json:"collectionName"`
}

type RegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
}

// --- Auth methods ---

func (c *Client) Login(ctx context.Context, email, password string) (*AuthResponse, error) {
	var resp AuthResponse
	err := c.post(ctx, "/collections/users/auth-with-password", map[string]string{
		"identity": email,
		"password": password,
	}, &resp)
	if err != nil {
		return nil, err
	}
	c.token = resp.Token
	return &resp, nil
}

func (c *Client) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	var resp AuthResponse
	err := c.post(ctx, "/collections/users/records", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) RefreshToken(ctx context.Context) (*AuthResponse, error) {
	var resp AuthResponse
	err := c.post(ctx, "/collections/users/auth-refresh", nil, &resp)
	if err != nil {
		return nil, err
	}
	c.token = resp.Token
	return &resp, nil
}

func (c *Client) Logout() {
	c.token = ""
}

func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	var resp AuthResponse
	err := c.post(ctx, "/collections/users/auth-refresh", nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Record, nil
}

func (c *Client) RequestPasswordReset(ctx context.Context, email string) error {
	return c.post(ctx, "/collections/users/request-password-reset", map[string]string{
		"email": email,
	}, nil)
}

func (c *Client) RequestVerification(ctx context.Context, email string) error {
	return c.post(ctx, "/collections/users/request-verification", map[string]string{
		"email": email,
	}, nil)
}

// --- Internal HTTP methods ---

func (c *Client) get(ctx context.Context, path string, result any) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

func (c *Client) post(ctx context.Context, path string, body, result any) error {
	return c.do(ctx, http.MethodPost, path, body, result)
}

func (c *Client) do(ctx context.Context, method, path string, body, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	fullURL, err := url.JoinPath(c.baseURL, c.prefix, path)
	if err != nil {
		return fmt.Errorf("failed to build url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		errBody, _ := io.ReadAll(resp.Body)
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       errBody,
		}
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

type APIError struct {
	StatusCode int
	Body       []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("auth api error %d: %s", e.StatusCode, string(e.Body))
}
