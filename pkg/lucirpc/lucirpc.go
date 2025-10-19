package lucirpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

const (
	rpcPath     = "/cgi-bin/luci/rpc/"
	authPath    = rpcPath + "auth"
	uciPath     = rpcPath + "uci"
	methodLogin = "login"

	defaultTimeout = 15
)

var (
	ErrRpcLoginFail        = errors.New("rpc: login fail")
	ErrHttpUnauthenticated = errors.New("http: Unauthenticated")
	ErrHttpUnauthorized    = errors.New("http: Unauthorized")
	ErrHttpForbidden       = errors.New("http: Forbidden")
)

// Payload represents a JSON-RPC request payload.
type Payload struct {
	ID     int      `json:"id"`
	Method string   `json:"method"`
	Params []string `json:"params"`
}

// Response represents a JSON-RPC response.
type Response struct {
	ID     int `json:"id"`
	Result any `json:"result"`
	Error  any `json:"error"`
}

// LuciRPC is a client for LuCI RPC API.
type LuciRPC struct {
	addr       string
	username   string
	password   string
	httpClient *http.Client
	rpcID      int
	token      string
}

// New creates a new LuciRPC client.
func New(addr, username, password string, rpcID int, insecureSkipVerify bool) (*LuciRPC, error) {
	if addr == "" {
		return nil, errors.New("address is empty")
	}

	if rpcID <= 0 {
		return nil, errors.New("rpcID must be greater than zero")
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecureSkipVerify,
			},
			Dial: (&net.Dialer{
				Timeout:   time.Duration(defaultTimeout) * time.Second,
				KeepAlive: time.Duration(defaultTimeout) * time.Second,
			}).Dial,
		},
	}

	return &LuciRPC{
		addr:       addr,
		username:   username,
		password:   password,
		httpClient: httpClient,
		rpcID:      rpcID,
	}, nil
}

// Uci performs a UCI RPC call with authentication.
func (c *LuciRPC) Uci(ctx context.Context, method string, params []string) (string, error) {
	return c.rpcWithAuth(ctx, uciPath, method, params)
}

func (c *LuciRPC) auth(ctx context.Context) error {
	token, err := c.rpc(ctx, authPath, methodLogin, []string{c.username, c.password})
	if err != nil {
		return err
	}

	// OpenWRT JSON RPC response of wrong username and password
	// {"id":1,"result":null,"error":null}
	if token == "null" {
		return ErrRpcLoginFail
	}

	c.token = token
	return nil
}

func (c *LuciRPC) rpc(ctx context.Context, path, method string, params []string) (string, error) {
	data, err := json.Marshal(Payload{
		ID:     c.rpcID,
		Method: method,
		Params: params,
	})
	if err != nil {
		return "", err
	}

	url := c.getUri(path, method)
	respBody, err := c.call(ctx, url, data)
	if err != nil {
		return "", err
	}

	var response Response
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", err
	}

	if response.Error != nil {
		return "", parseError(response.Error)
	}

	if response.Result != nil {
		return parseString(response.Result)
	}

	return "", nil
}

func (c *LuciRPC) getUri(path, method string) string {
	url := c.addr + path
	if method != methodLogin && c.token != "" {
		url = url + "?auth=" + c.token
	}

	return url
}

func (c *LuciRPC) call(ctx context.Context, url string, postBody []byte) ([]byte, error) {
	body := bytes.NewReader(postBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	if resp.StatusCode > 226 {
		return respBody, c.httpError(resp.StatusCode)
	}

	return respBody, err
}

func (c *LuciRPC) httpError(code int) error {
	if code == 401 {
		return ErrHttpUnauthorized
	}

	if code == 403 {
		return ErrHttpForbidden
	}

	return fmt.Errorf("http status code: %d", code)
}

func (c *LuciRPC) rpcWithAuth(ctx context.Context, path, method string, params []string) (string, error) {
	result, err := c.rpc(ctx, path, method, params)
	if err == nil {
		return result, nil
	}

	if err != ErrHttpUnauthorized && err != ErrHttpForbidden {
		return "", err
	}

	if err = c.auth(ctx); err != nil {
		return "", err
	}

	return c.rpc(ctx, path, method, params)
}

func parseString(obj any) (string, error) {
	if obj == nil {
		return "", errors.New("nil object cannot be parsed")
	}

	var result string
	if _, ok := obj.(string); ok {
		result = fmt.Sprintf("%v", obj)
		return result, nil
	}

	jsonBytes, err := json.Marshal(obj)
	if err == nil {
		result = string(jsonBytes)
	}

	return result, err
}

func parseError(obj any) error {
	result, err := parseString(obj)
	if err != nil {
		return err
	}

	return errors.New(result)
}
