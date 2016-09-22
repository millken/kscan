package icmp

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/millken/kscan/layer"
	"github.com/millken/kscan/program"
	"github.com/millken/kscan/server"
	"github.com/millken/kscan/utils"
)

type Boot struct {
	s   *server.Server
	pcf proconf
	dst []string
}

var cf Config
var pid = uint16(os.Getpid() & 0xffff)

func (t *Boot) Init(s *server.Server) (err error) {
	var srcMac, dstMac net.HardwareAddr
	var srcIP net.IP
	var body []byte
	t.s = s
	conf := s.ProgramConfig
	if _, ok := conf[NAME]; !ok {
		return fmt.Errorf("%s config is nil", NAME)
	}
	if err = toml.PrimitiveDecode(conf[NAME], &cf); err != nil {
		err = fmt.Errorf("Can't unmarshal config: %s", err)
	}
	srcMac, err = net.ParseMAC(cf.SrcMac)
	if err != nil {
		return
	}
	t.pcf.srcMAC = srcMac

	dstMac, err = net.ParseMAC(cf.DstMac)
	if err != nil {
		return
	}
	t.pcf.dstMAC = dstMac

	srcIP = net.ParseIP(cf.SrcIp)
	if srcIP == nil {
		return fmt.Errorf("ip format error :%s", srcIP)
	}
	t.pcf.srcIP = srcIP
	dstIP := net.ParseIP(cf.DstIp)
	if dstIP == nil {
		if body, err = utils.File_Get_Contents(cf.DstIp); err != nil {
			return fmt.Errorf("dst_ip not cidr or file not exist")
		} else {
			t.dst = strings.Split(string(body), "\n")
		}
	} else {
		t.dst = []string{cf.DstIp}
	}

	go t.readPackets()
	t.sendPackets()
	return nil
}

func onRecvPing(pkt gopacket.Packet, icmp *layers.ICMPv4) {
	payload := icmp.Payload
	if payload == nil || len(payload) <= 0 {
		return
	}
	sendStamp := binary.LittleEndian.Uint64(payload)
	if sendStamp < 1000000 {
		return
	}

	m := pkt.Metadata()

	delay := uint64(m.CaptureInfo.Timestamp.UnixNano()) - sendStamp
	ms := int(delay / 1000000)
	log.Printf("[DEBUG] ip = %s, %dms", layer.GetSrc(pkt), ms)
}
func (t *Boot) readPackets() {
	for {
		select {
		case pkt := <-t.s.RxChan:
			icmpLayer := pkt.Layer(layers.LayerTypeICMPv4)
			if icmpLayer == nil {
				return
			}
			icmp, _ := icmpLayer.(*layers.ICMPv4)
			offset := icmp.Seq & 0xff00
			switch {
			case icmp.Id == pid && offset == 0:
				onRecvPing(pkt, icmp)
			default:
			}
		}
	}
}

func (t *Boot) sendPackets() {
	p := layer.Packet{}
	p.Ethernet = &layers.Ethernet{
		SrcMAC:       t.pcf.srcMAC,
		DstMAC:       t.pcf.dstMAC,
		EthernetType: layers.EthernetTypeIPv4,
	}
	p.Ipv4 = &layers.IPv4{
		Version:  4,
		SrcIP:    t.pcf.srcIP,
		DstIP:    net.IP{42, 120, 60, 81},
		Protocol: layers.IPProtocolICMPv4,
		TTL:      64,
		Flags:    layers.IPv4DontFragment,
		Length:   20,
	}
	typeCode := layers.ICMPv4TypeEchoRequest
	p.Icmpv4 = &layers.ICMPv4{
		TypeCode: layers.ICMPv4TypeCode(uint16(typeCode) << 8),
		Seq:      1,
		Id:       pid,
		Checksum: 0,
	}

	p.Payload = make([]byte, 8)
	for _, ips := range t.dst {

		hosts, err := layer.Hosts(ips)
		if err == nil {
			for _, ip := range hosts {
				now := time.Now().UnixNano()
				binary.LittleEndian.PutUint64(p.Payload, uint64(now))
				p.Ipv4.DstIP = net.ParseIP(ip)
				if err := t.s.Send(p); err != nil {
					log.Printf("[ERROR] Send Packet : %s", err)
				}
			}
		}
	}

}

func init() {
	program.RegisterBooter(NAME, func() interface{} {
		return new(Boot)
	})
}
