package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTransferActualHTTPAndFailure(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("12345678")) }))
	defer s.Close()
	result := transfer(context.Background(), s.URL, "download", 8, nil, 2*time.Second)
	if !result.OK || result.Bytes != 8 || result.Mbps <= 0 {
		t.Fatalf("真实HTTP传输未成功: %+v", result)
	}
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(403) }))
	defer bad.Close()
	result = transfer(context.Background(), bad.URL, "download", 8, nil, time.Second)
	if result.OK || result.Error == "" {
		t.Fatal("HTTP错误必须失败")
	}
	result = transfer(context.Background(), s.URL, "download", 16, nil, time.Second)
	if result.OK {
		t.Fatal("不完整载荷不能算成功")
	}
}
