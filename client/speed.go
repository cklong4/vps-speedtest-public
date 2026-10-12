package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type SpeedResult struct {
	Direction string  `json:"direction"`
	Streams   int     `json:"streams"`
	OK        bool    `json:"ok"`
	Bytes     int64   `json:"bytes"`
	Seconds   float64 `json:"seconds"`
	Mbps      float64 `json:"mbps"`
	Error     string  `json:"error,omitempty"`
}

func directHTTP(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout, Transport: &http.Transport{Proxy: nil, DisableKeepAlives: true, DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		return (&net.Dialer{Timeout: 10 * time.Second}).DialContext(ctx, "tcp4", addr)
	}}}
}
func transfer(ctx context.Context, url, direction string, size int64, upload []byte, timeout time.Duration) SpeedResult {
	result := SpeedResult{Direction: direction, Streams: 1}
	begin := time.Now()
	var body io.Reader
	method := "GET"
	if direction == "upload" {
		method = "POST"
		body = bytes.NewReader(upload)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	req.Header.Set("User-Agent", "vps-speedtest-public/3.0")
	if method == "POST" {
		req.Header.Set("Content-Type", "application/octet-stream")
	}
	client := directHTTP(timeout)
	defer client.CloseIdleConnections()
	resp, err := client.Do(req)
	if err != nil {
		result.Seconds = time.Since(begin).Seconds()
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		result.Seconds = time.Since(begin).Seconds()
		return result
	}
	if direction == "download" {
		result.Bytes, err = io.Copy(io.Discard, io.LimitReader(resp.Body, size+1))
		if err == nil && result.Bytes != size {
			err = fmt.Errorf("载荷不完整: %d/%d", result.Bytes, size)
		}
	} else {
		_, err = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024*1024))
		if err == nil {
			result.Bytes = size
		}
	}
	result.Seconds = time.Since(begin).Seconds()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.OK = true
	result.Mbps = float64(result.Bytes) * 8 / result.Seconds / 1e6
	return result
}
func stage(ctx context.Context, endpoint, direction string, streams int, upload []byte, downloadSize int64) SpeedResult {
	n := downloadSize
	url := endpoint + fmt.Sprintf("/__down?bytes=%d", n)
	if direction == "upload" {
		n = int64(len(upload))
		url = endpoint + "/__up"
	}
	start := time.Now()
	ch := make(chan SpeedResult, streams)
	var wg sync.WaitGroup
	for i := 0; i < streams; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); ch <- transfer(ctx, url, direction, n, upload, 30*time.Second) }()
	}
	wg.Wait()
	close(ch)
	r := SpeedResult{Direction: direction, Streams: streams, OK: true, Seconds: time.Since(start).Seconds()}
	for x := range ch {
		r.Bytes += x.Bytes
		if !x.OK {
			r.OK = false
			r.Error += x.Error + "; "
		}
	}
	if r.OK {
		r.Mbps = float64(r.Bytes) * 8 / r.Seconds / 1e6
	}
	return r
}
func speedtest(ctx context.Context, endpoint string) ([]SpeedResult, error) {
	upload := make([]byte, 4*1024*1024)
	if _, err := rand.Read(upload); err != nil {
		return nil, err
	}
	results := []SpeedResult{}
	fmt.Println("[1/3] 本机速度测试：单/4连接下载、上传；计划载荷60MiB，不等于家宽到VPS速度")
	for _, dir := range []string{"download", "upload"} {
		for _, n := range []int{1, 4} {
			if ctx.Err() != nil {
				return results, ctx.Err()
			}
			r := stage(ctx, endpoint, dir, n, upload, 8*1024*1024)
			results = append(results, r)
			label := map[string]string{"download": "下载", "upload": "上传"}[dir]
			if r.OK {
				fmt.Printf("%s %d连接: %.2f Mbps (%.2f MB/s)\n", label, n, r.Mbps, r.Mbps/8)
			} else {
				fmt.Printf("%s %d连接: 失败 — %s\n", label, n, r.Error)
			}
		}
	}
	return results, nil
}
