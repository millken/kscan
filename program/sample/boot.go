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
}

var cf Config

func (t *Boot) Init(s *server.Server) (err error) {
	conf := s.ProgramConfig
	log.Printf("[DEBUG] sample.conf = %v", conf)
	if _, ok := conf[NAME]; !ok {
		return fmt.Errorf("%s config is nil", NAME)
	}
	if err = toml.PrimitiveDecode(conf[NAME], &cf); err != nil {
		err = fmt.Errorf("Can't unmarshal config: %s", err)
	}
	log.Printf("[DEBUG] sample = %v", cf.SrcMac)

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
		s.TxChan <- p
	}

	return nil
}

func init() {
	program.RegisterBooter(NAME, func() interface{} {
		return new(Boot)
	})
}
