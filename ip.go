package main

import (
	"encoding/binary"
	"math/rand"
	"net"
	"runtime"
	"syscall"
)

const IPV4HeaderLen = 20

func GetRandPublicIP() net.IP {

	for {
		ip := net.IP(make([]byte, 4))
		binary.BigEndian.PutUint32(ip[0:], uint32(rand.Intn(1<<32-1)))

		// we don't want private IPs
		if ip[0] == 10 || (ip[0] == 172 && ip[1] == 16) || (ip[0] == 192 && ip[1] == 168) || (ip[0] == 169 && ip[1] == 254) {
			continue
		}
		return ip
	}
}

// WriteIPV4Header writes a 20-byte IPv4 header into b[:IPV4HeaderLen].
// len(b) must equal the total packet length (header + payload).
func WriteIPV4Header(b []byte, srcIP, dstIP net.IP, totalLen int) {

	src := srcIP.To4()
	dst := dstIP.To4()

	b[0] = 0x45 // version 4, header length 20 bytes (5 * 4)
	b[1] = 0    // TOS

	// BSD-derived raw IP sockets (IP_HDRINCL) expect ip_len in host byte
	// order; everyone else expects wire (big-endian) byte order.
	switch runtime.GOOS {
	case "darwin", "ios", "dragonfly", "netbsd":
		binary.NativeEndian.PutUint16(b[2:4], uint16(totalLen))
	default:
		binary.BigEndian.PutUint16(b[2:4], uint16(totalLen))
	}
	binary.BigEndian.PutUint16(b[4:6], 1) // ID
	binary.BigEndian.PutUint16(b[6:8], 0) // flags + frag offset
	b[8] = 255                            // TTL
	b[9] = syscall.IPPROTO_TCP
	binary.BigEndian.PutUint16(b[10:12], 0) // checksum placeholder
	copy(b[12:16], src)
	copy(b[16:20], dst)

	binary.BigEndian.PutUint16(b[10:12], ipv4Checksum(b[:IPV4HeaderLen]))
}

func ipv4Checksum(data []byte) uint16 {

	var csum uint32
	for i := 0; i < len(data); i += 2 {
		csum += uint32(data[i])<<8 + uint32(data[i+1])
	}
	for csum > 0xffff {
		csum = (csum >> 16) + (csum & 0xffff)
	}
	return ^uint16(csum)
}
