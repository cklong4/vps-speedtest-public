package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const endpoint = "https://nezha.hzdbg.com/vps-api"

var version = "3.0.0"

type Session struct {
	ID     string `json:"id"`
	Access string `json:"access"`
	Status string `json:"status"`
}
type Job struct {
	ID            string                     `json:"id"`
	Status        string                     `json:"status"`
	Start         float64                    `json:"start"`
	End           float64                    `json:"end"`
	Report        string                     `json:"report"`
	ReportURL     string                     `json:"report_url"`
	ReportExpires float64                    `json:"report_expires"`
	Verified      bool                       `json:"verified_complete"`
	Error         string                     `json:"error"`
	Sources       map[string]json.RawMessage `json:"sources"`
}

func safeSave(path string, v interface{}) error {
	raw, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	file, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	defer file.Close()
	_, e = file.Write(raw)
	return e
}
func run() int {
	port := flag.Int("port", 22, "被测TCP端口，SSH不是22时请修改")
	resume := flag.String("resume", "", "恢复查看已有任务的JSON文件，不重开测试")
	local := flag.Bool("local-only", false, "仅本机速度与回程参考，不调用四来源")
	reportDir := flag.String("output", ".", "报告与本轮查看凭据保存目录")
	ver := flag.Bool("version", false, "版本")
	flag.Parse()
	if *ver {
		fmt.Println(version)
		return 0
	}
	if *port < 1 || *port > 65535 {
		fmt.Fprintln(os.Stderr, "端口必须1–65535")
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if e := os.MkdirAll(*reportDir, 0700); e != nil {
		fmt.Fprintln(os.Stderr, e)
		return 1
	}
	var session Session
	if *resume != "" {
		stat, e := os.Stat(*resume)
		if e != nil || stat.Mode().Perm()&0077 != 0 {
			fmt.Fprintln(os.Stderr, "任务文件不存在或权限不是0600")
			return 1
		}
		raw, e := os.ReadFile(*resume)
		if e != nil || json.Unmarshal(raw, &session) != nil {
			fmt.Fprintln(os.Stderr, "任务文件不可读")
			return 1
		}
	} else {
		fmt.Printf("VPS 一键网络测试 %s | 先速度，后线路参考与四来源稳定性\n", version)
		tests, e := speedtest(ctx, "https://speed.cloudflare.com")
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			return 1
		}
		traces := routes(ctx)
		preliminary := map[string]interface{}{"time": time.Now().UTC().Format(time.RFC3339), "speed": map[string]interface{}{"tests": tests}, "routes": traces}
		draft := filepath.Join(*reportDir, "vps-local-"+time.Now().Format("20060102-150405")+".json")
		if e := safeSave(draft, preliminary); e != nil {
			fmt.Fprintln(os.Stderr, "保存本机结果失败:", e)
			return 1
		}
		fmt.Println("本机原始结果:", draft)
		if *local {
			return 0
		}
		if ctx.Err() != nil {
			return 130
		}
		fmt.Println("[3/3] 四来源延迟/波动测试：匿名只测本机公网IPv4，固定10分钟；不会使用私有节点账号")
		var info struct {
			IP string `json:"your_ip"`
		}
		if e := api(ctx, endpoint, "/info", nil, "", &info); e != nil {
			fmt.Fprintln(os.Stderr, e)
			return 1
		}
		fmt.Printf("协调端核对的本机IP: %s | TCP端口: %d\n", info.IP, *port)
		request := map[string]interface{}{"target": info.IP, "port": *port, "minutes": 10, "speed": map[string]interface{}{"tests": tests}, "routes": traces}
		if e := api(ctx, endpoint, "/jobs", request, "", &session); e != nil {
			fmt.Fprintln(os.Stderr, "未创建四来源任务:", e, "；本机速度结果已保留")
			return 1
		}
		path := filepath.Join(*reportDir, ".vps-network-job-"+session.ID+".json")
		if e := safeSave(path, session); e != nil {
			fmt.Fprintln(os.Stderr, "保存恢复文件失败:", e)
			return 1
		}
		fmt.Println("本轮恢复文件（0600，不是API Key）:", path)
	}
	last := ""
	announced := false
	failedReads := 0
	for ctx.Err() == nil {
		var job Job
		if e := api(ctx, endpoint, "/jobs/"+session.ID, nil, session.Access, &job); e != nil {
			failedReads++
			if failedReads >= 3 {
				fmt.Fprintln(os.Stderr, "连续读取失败，任务按截止收尾，请稍后--resume恢复:", e)
				return 1
			}
		} else {
			failedReads = 0
			if job.Status != last {
				fmt.Println("状态:", job.Status)
				last = job.Status
			}
			if !announced && (job.Status == "running" || job.Status == "finishing") {
				fmt.Printf("已读回四来源真实样本，共同窗口 %s 至 %s\n", time.Unix(int64(job.Start), 0).Format(time.RFC3339), time.Unix(int64(job.End), 0).Format(time.RFC3339))
				announced = true
			}
			if job.Status == "completed" || job.Status == "incomplete" || job.Status == "failed" || job.Status == "interrupted" {
				fmt.Println(job.Report)
				file := filepath.Join(*reportDir, "vps-network-"+job.ID+".txt")
				if e := os.WriteFile(file, []byte(job.Report), 0600); e != nil {
					fmt.Fprintln(os.Stderr, "本地报告保存失败:", e)
				} else {
					fmt.Println("完整报告:", file)
				}
				fmt.Println("7天报告链接:", job.ReportURL)
				fmt.Println("到期时间:", time.Unix(int64(job.ReportExpires), 0).Format(time.RFC3339))
				fmt.Println("拿到链接的人都能看；到期自动失效；本机速度为客户端自测，非协调端独立校验")
				if job.Verified {
					return 0
				}
				return 1
			}
		}
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
		}
	}
	fmt.Println("已停止终端显示；已创建的服务端任务仍按10分钟截止收尾，不会重开。用恢复文件查看同一轮。")
	return 130
}
func main() { os.Exit(run()) }
