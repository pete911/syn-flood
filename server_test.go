//go:build manual

package main

import (
	"log"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestStartServer starts a plain TCP listener you can point the syn-flood
// CLI at manually from another terminal, e.g.:
//
//	go test -tags manual -run TestStartServer -v ./...   # start the server
//	./syn-flood -host 127.0.0.1 -port 9999                # in another terminal
//
// It just blocks, periodically printing netstat's view of the connection
// table (spoofed SYNs never complete a handshake, so Accept() alone won't
// show them), rather than asserting anything. It's gated behind the
// "manual" build tag and excluded from `go test ./...`. Stop it with Ctrl+C.
func TestStartServer(t *testing.T) {

	ln, err := net.Listen("tcp", "127.0.0.1:9999")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	log.Printf("listening on %s - point syn-flood at this host:port", ln.Addr())

	go func() {
		for {
			time.Sleep(2 * time.Second)

			synRecv := execCmd("bash", "-c", "netstat -an -p tcp | grep 9999 | grep SYN_RCVD | wc -l")
			log.Printf("SYN_RCVD: %s", synRecv)

			if err := probeServer(); err != nil {
				log.Printf("client request failed: %v", err)
			} else {
				log.Println("client request succeeded")
			}
		}
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			t.Logf("accept: %v", err)
			return
		}
		log.Printf("accepted connection from %s", conn.RemoteAddr())
		conn.Close()
	}
}

func probeServer() error {

	conn, err := net.DialTimeout("tcp", "127.0.0.1:9999", time.Second)
	if err != nil {
		return err
	}
	return conn.Close()
}

func execCmd(name string, arg ...string) string {
	b, err := exec.Command(name, arg...).Output()
	if err != nil {
		return err.Error()
	}
	return strings.TrimSpace(string(b))
}
