package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/millken/kscan/config"
	"github.com/millken/kscan/drivers"
	"github.com/millken/kscan/layer"
	"github.com/millken/kscan/stats"
)

const BPFFilter = "udp"

type server struct {
	config        *config.Config
	io            drivers.PacketDataSourceCloser
	stats         stats.Directional
	isStopped     bool
	forceQuitChan chan os.Signal
	txChan        chan layer.Packet
	rxChan        chan gopacket.Packet
}

func New(cf *config.Config) *server {
	return &server{
		config:        cf,
		forceQuitChan: make(chan os.Signal, 1),
		txChan:        make(chan layer.Packet),
		rxChan:        make(chan gopacket.Packet),
		isStopped:     false,
	}
}

func (s *server) Start() (err error) {
	options := &drivers.DriverOptions{
		Device:  s.config.Server.Iface,
		Snaplen: 2048,
		Filter:  BPFFilter,
	}

	factory, ok := drivers.Drivers[s.config.Server.Driver]
	if !ok {
		log.Fatal(fmt.Sprintf("%s Packet driver not supported on this system", s.config.Server.Driver))
	}

	s.io, err = factory(options)
	if err != nil {
		return fmt.Errorf("driver: %s, interface: %s boot error: %s", s.config.Server.Driver, s.config.Server.Iface, err)
	}

	//go s.readPackets()
	s.sendPackets()

	worker_num := 7
	if s.config.Server.WorkerNum > 0 {
		worker_num = s.config.Server.WorkerNum
	}
	for i := 0; i < worker_num; i++ {
		//go packetHandler(i, s.rxChan, s.txChan)
	}
	//s.signalWorker()
	return
}

func (s *server) sendPackets() {
	var err error
	defer close(s.txChan)

	p := layer.Packet{}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	p.Ethernet = &layers.Ethernet{
		SrcMAC:       net.HardwareAddr{0x00, 0x1B, 0x21, 0xA7, 0x9A, 0x34},
		DstMAC:       net.HardwareAddr{0x00, 0x1B, 0x21, 0x98, 0xE8, 0x8C},
		EthernetType: layers.EthernetTypeIPv4,
	}
	p.Ipv4 = &layers.IPv4{
		SrcIP:    net.IP{192, 168, 5, 84},
		DstIP:    net.IP{192, 168, 5, 85},
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
	out := make([]byte, 0, 64)

	p.Udp.SetNetworkLayerForChecksum(p.Ipv4)
	gopacket.SerializeLayers(buf, opts, p.Ethernet, p.Ipv4, p.Udp, gopacket.Payload(out))

	start := time.Now()
	for i := 0; i < 1000000; i++ {
		err = s.io.WritePacketData(buf.Bytes())
		//	if err != nil {
		//	log.Fatalf("write packet data err : %+v", err)
		//}

		//s.stats.Tx.Packets++
		//d.stats.Tx.Bytes += uint64(p.Metadata().CaptureInfo.CaptureLength)
	}
	end := time.Now()
	for i := 0; i < 10000000; i++ {
	}
	end1 := time.Now()
	log.Printf("[INFO] tx pkts = %d, time: %v, %v, err:%s", s.stats.Tx.Packets, end.Sub(start), end1.Sub(end), err)
}

func (s *server) readPackets() {
	packetSource := s.io.PacketSource()
	defer close(s.rxChan)
	for {

		packet, err := packetSource.NextPacket()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("[ERROR] readPackets err: %s", err)
			continue
		}
		s.stats.Rx.Packets++

		s.rxChan <- packet
		if s.isStopped {
			break
		}
	}
}

func (s *server) Shutdown() {
	s.io.Close()
	s.isStopped = true
}

func (s *server) Stats() stats.Directional {
	return s.stats
}

func (s *server) signalWorker() {
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM,
		syscall.SIGINT)

	for {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP:
			log.Println("Reload initiated.")
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			s.Shutdown()
			log.Println("Shutdown initiated.")
			return
		}
	}
}
