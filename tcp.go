package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"syscall"
)

const TCPSYNHeaderLen = 20

// ephemeral-range random source port, avoids the syscall overhead of
// binding a real socket just to obtain a free port number.
func getSrcTCPPort() uint16 {
	return uint16(1024 + rand.Intn(65535-1024))
}

// WriteTCPSYNHeader writes a 20-byte TCP SYN header into b[:TCPSYNHeaderLen].
func WriteTCPSYNHeader(b []byte, srcIP, dstIP net.IP, dstPort uint16) error {

	srcPort := getSrcTCPPort()
	seq := rand.Intn(1<<32 - 1)

	binary.BigEndian.PutUint16(b[0:2], srcPort)     // source port
	binary.BigEndian.PutUint16(b[2:4], dstPort)     // destination port
	binary.BigEndian.PutUint32(b[4:8], uint32(seq)) // sequence number
	binary.BigEndian.PutUint32(b[8:12], 0)          // acknowledgement number
	b[12] = TCPSYNHeaderLen / 4 << 4                // data offset
	b[13] = 0b00000010                              // only SYN flag set
	binary.BigEndian.PutUint16(b[14:16], 65535)     // window size
	binary.BigEndian.PutUint16(b[16:18], 0)         // checksum placeholder
	binary.BigEndian.PutUint16(b[18:20], 0)         // urgent pointer

	checksum, err := tcpChecksum(srcIP, dstIP, b[:TCPSYNHeaderLen])
	if err != nil {
		return fmt.Errorf("tcp checksum: %w", err)
	}
	binary.BigEndian.PutUint16(b[16:18], checksum)
	return nil
}

func tcpChecksum(srcIP, dstIP net.IP, data []byte) (uint16, error) {

	src := srcIP.To4()
	if src == nil {
		return 0, fmt.Errorf("src IP: not an IPv4 address")
	}
	dst := dstIP.To4()
	if dst == nil {
		return 0, fmt.Errorf("dst IP: not an IPv4 address")
	}

	// pseudo-header: src address, dst address, zero byte + protocol, TCP length
	var csum uint32
	csum += uint32(src[0])<<8 + uint32(src[1])
	csum += uint32(src[2])<<8 + uint32(src[3])
	csum += uint32(dst[0])<<8 + uint32(dst[1])
	csum += uint32(dst[2])<<8 + uint32(dst[3])
	csum += uint32(syscall.IPPROTO_TCP)
	csum += uint32(len(data))

	// to handle odd lengths, we loop to length - 1, incrementing by 2, then
	// handle the last byte specifically by checking against the original
	// length.
	length := len(data) - 1
	for i := 0; i < length; i += 2 {
		csum += uint32(data[i])<<8 + uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		csum += uint32(data[length]) << 8
	}
	for csum > 0xffff {
		csum = (csum >> 16) + (csum & 0xffff)
	}
	return ^uint16(csum), nil
}
