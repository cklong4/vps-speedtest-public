package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var apiClient = directHTTP(40 * time.Second)

func api(ctx context.Context, base, path string, data interface{}, access string, out interface{}) error {
	if !strings.HasPrefix(base, "https://") {
		return fmt.Errorf("协调端必须HTTPS，不允许明文降级")
	}
	method := "GET"
	var body io.Reader
	if data != nil {
		raw, err := json.Marshal(data)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
		method = "POST"
	}
	req, err := http.NewRequestWithContext(ctx, method, base+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "vps-speedtest-public/3.0")
	if access != "" {
		req.Header.Set("X-Job-Token", access)
	}
	response, err := apiClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(response.Body, 4*1024*1024))
	if err != nil {
		return err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		var e struct {
			Error string `json:"error"`
		}
		json.Unmarshal(raw, &e)
		if e.Error == "" {
			e.Error = string(raw)
			if len(e.Error) > 180 {
				e.Error = e.Error[:180]
			}
		}
		return fmt.Errorf("HTTP %d: %s", response.StatusCode, e.Error)
	}
	return json.Unmarshal(raw, out)
}
