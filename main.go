package main

import (
	"log"
	"math/rand"
	"net"
	"syscall"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	srcIp := GetRandPublicIP()
	// TODO make dst ip and port as args/flags
	dstIp := net.ParseIP("127.0.0.1")
	dstPort := 80

	tcpHeaderBytes, err := GetTCPSYNHeaderBytes(srcIp, dstIp, uint16(dstPort))
	if err != nil {
		log.Fatalf("get TCP header: %v", err)
	}

	ipv4Header := GetIPV4Header(srcIp, dstIp, len(tcpHeaderBytes), syscall.IPPROTO_TCP)
	ipv4HeaderBytes, _ := ipv4Header.Marshal()

	data := append(ipv4HeaderBytes[:], tcpHeaderBytes[:]...)
	SendTo(dstIp, data)
}

func SendTo(dstIP net.IP, data []byte) {

	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		panic(err)
	}

	err = syscall.SetsockoptInt(s, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	if err != nil {
		panic(err)
	}

	var addr syscall.SockaddrInet4
	copy(addr.Addr[:4], dstIP.To4())

	err = syscall.Sendto(s, data, 0, &addr)
	if err != nil {
		log.Fatalf("send to: %v", err)
	}
}
