package sample

import (
	"fmt"
	"log"
	"net"

	"github.com/BurntSushi/toml"
	"github.com/google/gopacket/layers"
	"github.com/millken/kscan/layer"
	"github.com/millken/kscan/program"
	"github.com/millken/kscan/server"
)

type Boot struct {
	s *server.Server
}

var cf Config

func (t *Boot) Init(s *server.Server) (err error) {
	t.s = s
	conf := s.ProgramConfig
	if _, ok := conf[NAME]; !ok {
		return fmt.Errorf("%s config is nil", NAME)
	}
	if err = toml.PrimitiveDecode(conf[NAME], &cf); err != nil {
		err = fmt.Errorf("Can't unmarshal config: %s", err)
	}
	go t.readPackets()
	t.SynTest()
	return nil
}

func (t *Boot) readPackets() {
	for {
		select {
		case pkt := <-t.s.RxChan:
			tcp := layer.GetTcp(pkt)
			if tcp.ACK && tcp.DstPort == 4064 {
				log.Printf("[DEBUG] ip = %s", layer.GetSrc(pkt))
			}
		}
	}
}

func (t *Boot) SynTest() {
	p := layer.Packet{}
	p.Ethernet = &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x50, 0x46, 0x5d, 0x59, 0xcb, 0xc5},
		DstMAC:       net.HardwareAddr{0x50, 0xda, 0x00, 0x4a, 0xab, 0x2f},
		EthernetType: layers.EthernetTypeIPv4,
	}
	p.Ipv4 = &layers.IPv4{
		Version:  4,
		SrcIP:    net.IP{192, 168, 5, 106},
		DstIP:    net.IP{42, 120, 60, 81},
		Protocol: layers.IPProtocolTCP,
		TTL:      64,
		Flags:    layers.IPv4DontFragment,
		Id:       19749,
	}
	p.Tcp = &layers.TCP{
		SrcPort: 4064,
		DstPort: 80,
		Window:  29200,
		SYN:     true,
	}
	p.Tcp.Options = append(p.Tcp.Options, layers.TCPOption{
		OptionType:   1, //TCPOptionKindNop
		OptionLength: 1,
	})

	hosts, _ := layer.Hosts("42.120.60.0/24")
	for _, ip := range hosts {
		p.Ipv4.DstIP = net.ParseIP(ip)
		t.s.TxChan <- p
	}
}

func (t *Boot) UdpTest() {
	p := layer.Packet{}
	p.Ethernet = &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x1B, 0x21, 0x99, 0x2A, 0x05},
		DstMAC:       net.HardwareAddr{0x00, 0x1B, 0x21, 0x99, 0x2A, 0x04},
		EthernetType: layers.EthernetTypeIPv4,
	}
	p.Ipv4 = &layers.IPv4{
		SrcIP:    net.IP{192, 168, 55, 99},
		DstIP:    net.IP{192, 168, 55, 100},
		Protocol: layers.IPProtocolUDP,
		TTL:      64,
		IHL:      5,
		Flags:    layers.IPv4DontFragment,
		Id:       964,
	}
	p.Udp = &layers.UDP{
		SrcPort: 41781,
		DstPort: 33434,
	}
	p.Payload = make([]byte, 0, 64)
	for i := 0; i < 100; i++ {
		t.s.TxChan <- p
	}
}

func init() {
	program.RegisterBooter(NAME, func() interface{} {
		return new(Boot)
	})
}
