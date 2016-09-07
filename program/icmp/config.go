package icmp

import "net"

const NAME = "icmp"

type Config struct {
	SrcMac string `toml:"src_mac"`
	DstMac string `toml:"dst_mac"`
	SrcIp  string `toml:"src_ip"`
	DstIp  string `toml:"dst_ip"`
}

type proconf struct {
	srcMAC net.HardwareAddr
	dstMAC net.HardwareAddr
	srcIP  net.IP
}
