package main

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestGetSrcTCPPort(t *testing.T) {

	for range 10000 {
		port := getSrcTCPPort()
		if port < 1024 {
			t.Fatalf("port %d below ephemeral range", port)
		}
	}
}

func TestWriteTCPSYNHeader(t *testing.T) {

	srcIP := net.ParseIP("1.2.3.4")
	dstIP := net.ParseIP("5.6.7.8")
	const dstPort = 443

	b := make([]byte, TCPSYNHeaderLen)
	if err := WriteTCPSYNHeader(b, srcIP, dstIP, dstPort); err != nil {
		t.Fatalf("WriteTCPSYNHeader: %v", err)
	}

	if got := binary.BigEndian.Uint16(b[2:4]); got != dstPort {
		t.Errorf("dst port = %d, want %d", got, dstPort)
	}
	if got, want := b[12]>>4, byte(TCPSYNHeaderLen/4); got != want {
		t.Errorf("data offset = %d words, want %d", got, want)
	}
	if got, want := b[13], byte(0b00000010); got != want {
		t.Errorf("flags = %#b, want SYN only (%#b)", got, want)
	}
	if got := binary.BigEndian.Uint16(b[14:16]); got != 65535 {
		t.Errorf("window = %d, want 65535", got)
	}

	// re-checksumming a header that already contains its own checksum must
	// yield 0 - this is the standard TCP checksum self-consistency property.
	checksum, err := tcpChecksum(srcIP, dstIP, b)
	if err != nil {
		t.Fatalf("tcpChecksum: %v", err)
	}
	if checksum != 0 {
		t.Errorf("checksum over header with checksum field populated = %#x, want 0", checksum)
	}
}

func TestWriteTCPSYNHeader_InvalidIP(t *testing.T) {

	b := make([]byte, TCPSYNHeaderLen)
	if err := WriteTCPSYNHeader(b, net.ParseIP("::1"), net.ParseIP("5.6.7.8"), 443); err == nil {
		t.Fatal("expected error for non-IPv4 source address")
	}
}
