package main

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestICMPOnlyMatchesOwnProbeDestinationAndPort(t *testing.T) {
	packet := make([]byte, 36)
	packet[0] = 11
	packet[8] = 0x45
	packet[17] = 17
	copy(packet[24:28], net.ParseIP("8.8.8.8").To4())
	binary.BigEndian.PutUint16(packet[30:32], 33440)
	if !icmpMatches(packet, net.ParseIP("8.8.8.8"), 33440) {
		t.Fatal("合法TTL超时未匹配")
	}
	if icmpMatches(packet, net.ParseIP("9.9.9.9"), 33440) || icmpMatches(packet, net.ParseIP("8.8.8.8"), 33441) {
		t.Fatal("误匹配其他探测")
	}
}
