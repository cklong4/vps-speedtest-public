package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"syscall"
	"time"
)

type Hop struct {
	TTL int     `json:"ttl"`
	IP  string  `json:"ip"`
	Ms  float64 `json:"ms"`
}
type Route struct {
	Target string `json:"target"`
	Label  string `json:"label"`
	Hops   []Hop  `json:"hops"`
	Error  string `json:"error,omitempty"`
}

func icmpMatches(packet []byte, target net.IP, port int) bool {
	if len(packet) < 36 || (packet[0] != 11 && packet[0] != 3) {
		return false
	}
	inner := packet[8:]
	n := int(inner[0]&15) * 4
	return n >= 20 && len(inner) >= n+8 && inner[9] == 17 && net.IP(inner[16:20]).Equal(target.To4()) && int(binary.BigEndian.Uint16(inner[n+2:n+4])) == port
}
func trace(ctx context.Context, target, label string) Route {
	result := Route{Target: target, Label: label, Hops: []Hop{}}
	listener, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		result.Error = "没有原始ICMP权限，回程路径未测: " + err.Error()
		return result
	}
	defer listener.Close()
	dst := net.ParseIP(target).To4()
	buf := make([]byte, 1500)
	base := 33434 + int(time.Now().UnixNano()%20000)
	for ttl := 1; ttl <= 24 && ctx.Err() == nil; ttl++ {
		port := base + ttl
		conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: dst, Port: port})
		if err != nil {
			result.Error = err.Error()
			break
		}
		raw, err := conn.SyscallConn()
		if err == nil {
			var socketErr error
			err = raw.Control(func(fd uintptr) { socketErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TTL, ttl) })
			if err == nil {
				err = socketErr
			}
		}
		if err != nil {
			conn.Close()
			result.Error = err.Error()
			break
		}
		begin := time.Now()
		_, err = conn.Write([]byte("vps-speedtest-route"))
		hop := Hop{TTL: ttl}
		reached := false
		if err == nil {
			until := begin.Add(650 * time.Millisecond)
			for time.Now().Before(until) {
				listener.SetReadDeadline(until)
				n, addr, e := listener.ReadFrom(buf)
				if e != nil {
					break
				}
				if icmpMatches(buf[:n], dst, port) {
					hop.IP = addr.String()
					hop.Ms = float64(time.Since(begin).Microseconds()) / 1000
					reached = net.ParseIP(hop.IP).Equal(dst)
					break
				}
			}
		}
		conn.Close()
		result.Hops = append(result.Hops, hop)
		if reached {
			break
		}
	}
	return result
}
func routes(ctx context.Context) []Route {
	fmt.Println("[2/3] 线路参考：VPS → 电信/联通/移动公网参考目标；看不到的跳点记未知，非产品等级证明")
	targets := [][2]string{{"202.101.172.35", "电信参考（AS4134）"}, {"210.22.70.3", "联通参考（AS17621）"}, {"211.136.192.6", "移动参考（AS56040）"}}
	out := []Route{}
	for _, t := range targets {
		if ctx.Err() != nil {
			break
		}
		r := trace(ctx, t[0], t[1])
		out = append(out, r)
		count := 0
		for _, h := range r.Hops {
			if h.IP != "" {
				count++
			}
		}
		fmt.Printf("%s: %d个应答跳点，%d跳尝试 %s\n", t[1], count, len(r.Hops), r.Error)
	}
	return out
}
