package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
)

const TCPSYNHeaderLen = 20

func getSrcTCPPort() (int, error) {

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func GetTCPSYNHeaderBytes(srcIP, dstIP net.IP, dstPort uint16) ([]byte, error) {

	srcPort, err := getSrcTCPPort()
	if err != nil {
		return nil, fmt.Errorf("get src tcp port: %w", err)
	}

	seq := rand.Intn(1<<32 - 1)
	offsetAndFlags := []byte{
		byte(TCPSYNHeaderLen / 4 << 4), // first 4 bits is data offset, tcp header length in 32 bits words (header length / 4)
		byte(0b00000010),               // only SYN flag set
	}

	b := make([]byte, TCPSYNHeaderLen)
	binary.BigEndian.PutUint16(b[0:2], uint16(srcPort)) // source port
	binary.BigEndian.PutUint16(b[2:4], dstPort)         // destination port
	binary.BigEndian.PutUint32(b[4:8], uint32(seq))     // sequence number
	binary.BigEndian.PutUint32(b[8:12], 0)              // acknowledgement number
	copy(b[12:14], offsetAndFlags)                      // offset, reserved and flags
	binary.BigEndian.PutUint16(b[14:16], 65535)         // window size
	binary.BigEndian.PutUint16(b[18:20], 0)             // urgent pointer

	checksum, err := tcpChecksum(srcIP, dstIP, b)
	if err != nil {
		return nil, fmt.Errorf("tcp checksum: %w", err)
	}
	binary.BigEndian.PutUint16(b[16:18], checksum)
	return b, nil
}

func tcpChecksum(srcIP, dstIP net.IP, data []byte) (uint16, error) {

	src, err := srcIP.To4().MarshalText()
	if err != nil {
		return 0, fmt.Errorf("src IP: %w", err)
	}
	dst, err := dstIP.To4().MarshalText()
	if err != nil {
		return 0, fmt.Errorf("dst IP: %w", err)
	}

	var csum uint32
	csum += (uint32(src[0]) + uint32(src[2])) << 8
	csum += uint32(src[1]) + uint32(src[3])
	csum += (uint32(dst[0]) + uint32(dst[2])) << 8
	csum += uint32(dst[1]) + uint32(dst[3])

	// to handle odd lengths, we loop to length - 1, incrementing by 2, then
	// handle the last byte specifically by checking against the original
	// length.
	length := TCPSYNHeaderLen - 1
	for i := 0; i < length; i += 2 {
		// For our test packet, doing this manually is about 25% faster
		// (740 ns vs. 1000ns) than doing it by calling binary.BigEndian.Uint16.
		csum += uint32(data[i]) << 8
		csum += uint32(data[i+1])
	}
	if len(data)%2 == 1 {
		csum += uint32(data[length]) << 8
	}
	for csum > 0xffff {
		csum = (csum >> 16) + (csum & 0xffff)
	}
	return ^uint16(csum), nil
}
