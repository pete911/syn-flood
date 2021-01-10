package main

import (
	"net"
	"syscall"
)

type RawSocket struct {
	fd   int
	addr syscall.SockaddrInet4
}

func NewRawSocket(dstIP net.IP) (RawSocket, error) {

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		syscall.Close(fd)
		return RawSocket{}, err
	}

	if err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1); err != nil {
		syscall.Close(fd)
		return RawSocket{}, err
	}

	var addr syscall.SockaddrInet4
	copy(addr.Addr[:4], dstIP.To4())
	return RawSocket{fd: fd, addr: addr}, nil
}

func (r RawSocket) Send(data []byte) error {
	return syscall.Sendto(r.fd, data, 0, &r.addr)
}
