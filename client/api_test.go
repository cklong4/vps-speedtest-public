package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAPIRefusesDowngradeAndReturnsServerQuota(t *testing.T) {
	var out map[string]interface{}
	if err := api(context.Background(), "http://example.com", "/info", nil, "", &out); err == nil {
		t.Fatal("不能明文传任务凭据")
	}
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		json.NewEncoder(w).Encode(map[string]string{"error": "每小时最多一次"})
	}))
	defer s.Close()
	old := apiClient
	apiClient = s.Client()
	defer func() { apiClient = old }()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := api(ctx, s.URL, "/jobs", map[string]string{}, "", &out)
	if err == nil || !strings.Contains(err.Error(), "每小时最多一次") {
		t.Fatal("配额错误未明确返回", err)
	}
}
