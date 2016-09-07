package layer

import (
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type Packet struct {
	Ethernet *layers.Ethernet
	Ipv4     *layers.IPv4
	Tcp      *layers.TCP
	Udp      *layers.UDP
	Icmpv4   *layers.ICMPv4
	Payload  []byte
}

func GetSrc(packet gopacket.Packet) net.IP {
	var ipv4 layers.IPv4

	ipv4_layer := packet.NetworkLayer()
	ipv4.DecodeFromBytes(ipv4_layer.LayerContents(), gopacket.NilDecodeFeedback)
	return ipv4.SrcIP
}

func GetTcp(packet gopacket.Packet) layers.TCP {
	var tcp layers.TCP

	tcp_layer := packet.TransportLayer()
	tcp.DecodeFromBytes(tcp_layer.LayerContents(), gopacket.NilDecodeFeedback)
	return tcp
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func Hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}
