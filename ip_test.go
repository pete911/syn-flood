package main

import (
	"encoding/binary"
	"net"
	"runtime"
	"testing"
)

func TestGetRandPublicIP(t *testing.T) {

	for i := 0; i < 10000; i++ {
		ip := GetRandPublicIP()
		if len(ip) != 4 {
			t.Fatalf("expected 4-byte IPv4 address, got %d bytes", len(ip))
		}
		if ip[0] == 10 {
			t.Fatalf("got private IP in 10.0.0.0/8: %s", ip)
		}
		if ip[0] == 172 && ip[1] == 16 {
			t.Fatalf("got private IP in 172.16.0.0/12: %s", ip)
		}
		if ip[0] == 192 && ip[1] == 168 {
			t.Fatalf("got private IP in 192.168.0.0/16: %s", ip)
		}
		if ip[0] == 169 && ip[1] == 254 {
			t.Fatalf("got link-local IP in 169.254.0.0/16: %s", ip)
		}
	}
}

func TestWriteIPV4Header(t *testing.T) {

	srcIP := net.ParseIP("1.2.3.4")
	dstIP := net.ParseIP("5.6.7.8")
	const totalLen = 40

	b := make([]byte, IPV4HeaderLen)
	WriteIPV4Header(b, srcIP, dstIP, totalLen)

	if got, want := b[0], byte(0x45); got != want {
		t.Errorf("version/IHL byte = %#x, want %#x", got, want)
	}
	if got, want := b[8], byte(255); got != want {
		t.Errorf("TTL = %d, want %d", got, want)
	}
	if got, want := b[9], byte(6); got != want { // syscall.IPPROTO_TCP
		t.Errorf("protocol = %d, want %d", got, want)
	}
	if !net.IP(b[12:16]).Equal(srcIP.To4()) {
		t.Errorf("src address = %v, want %v", net.IP(b[12:16]), srcIP)
	}
	if !net.IP(b[16:20]).Equal(dstIP.To4()) {
		t.Errorf("dst address = %v, want %v", net.IP(b[16:20]), dstIP)
	}

	// total length is byte-order dependent on BSD-derived kernels when
	// IP_HDRINCL is used - assert against the same rule the code applies.
	var gotTotalLen uint16
	switch runtime.GOOS {
	case "darwin", "ios", "dragonfly", "netbsd":
		gotTotalLen = binary.NativeEndian.Uint16(b[2:4])
	default:
		gotTotalLen = binary.BigEndian.Uint16(b[2:4])
	}
	if gotTotalLen != totalLen {
		t.Errorf("total length = %d, want %d", gotTotalLen, totalLen)
	}

	// re-checksumming a header that already contains its own checksum must
	// yield 0 - this is the standard IP checksum self-consistency property.
	if checksum := ipv4Checksum(b); checksum != 0 {
		t.Errorf("checksum over header with checksum field populated = %#x, want 0", checksum)
	}
}
