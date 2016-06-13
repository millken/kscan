package layer

import "github.com/google/gopacket/layers"

type Packet struct {
	Ethernet *layers.Ethernet
	Ipv4     *layers.IPv4
	Tcp      *layers.TCP
	Udp      *layers.UDP
}
