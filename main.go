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

	//srcIp := GetRandPublicIP()
	// 192.168.86.183
	srcIp := net.ParseIP("192.168.86.183")
	//srcIp := net.ParseIP("0.0.0.0")
	// TODO make dst ip and port as args/flags
	dstIp := net.ParseIP("127.0.0.1")
	//dstIp := net.ParseIP("192.168.86.116")
	dstPort := 8080

	tcpHeaderBytes, err := GetTCPSYNHeaderBytes(srcIp, dstIp, uint16(dstPort))
	if err != nil {
		log.Fatalf("get TCP header: %v", err)
	}

	ipv4Header := GetIPV4Header(srcIp, dstIp, len(tcpHeaderBytes), syscall.IPPROTO_TCP)
	ipv4HeaderBytes, _ := ipv4Header.Marshal()

	data := append(ipv4HeaderBytes[:], tcpHeaderBytes[:]...)

	rawSocket, err := NewRawSocket(dstIp)
	if err != nil {
		log.Fatalf("new raw socket: %v", err)
	}
	rawSocket.Send(data)
	time.Sleep(5 * time.Second)
}
