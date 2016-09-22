package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/gopacket"
	"github.com/millken/kscan/config"
	"github.com/millken/kscan/drivers"
	"github.com/millken/kscan/layer"
	"github.com/millken/kscan/stats"
)

type Server struct {
	Config        *config.MasterConf
	ProgramConfig config.ProgramConf
	io            drivers.PacketDataSourceCloser
	stats         stats.Directional
	isStopped     bool
	forceQuitChan chan os.Signal
	TxChan        chan layer.Packet
	RxChan        chan gopacket.Packet
}

func New(cf *config.MasterConf, pcf config.ProgramConf) *Server {
	return &Server{
		Config:        cf,
		ProgramConfig: pcf,
		forceQuitChan: make(chan os.Signal, 1),
		TxChan:        make(chan layer.Packet),
		RxChan:        make(chan gopacket.Packet),
		isStopped:     false,
	}
}

func (s *Server) Start() (err error) {
	options := &drivers.DriverOptions{
		Device:  s.Config.Iface,
		Snaplen: 2048,
		Filter:  s.Config.Filter,
	}

	factory, ok := drivers.Drivers[s.Config.Driver]
	if !ok {
		log.Fatal(fmt.Sprintf("%s Packet driver not supported on this system", s.Config.Driver))
	}

	s.io, err = factory(options)
	if err != nil {
		return fmt.Errorf("driver: %s, interface: %s boot error: %s", s.Config.Driver, s.Config.Iface, err)
	}

	go s.readPackets()
	go s.sendPackets()

	//go packetHandler(i, s.rxChan, s.txChan)
	//s.signalWorker()
	return
}

func (s *Server) Send(pkt layer.Packet) (err error) {
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	if pkt.Udp != nil {
		pkt.Udp.SetNetworkLayerForChecksum(pkt.Ipv4)
		gopacket.SerializeLayers(buf, opts, pkt.Ethernet, pkt.Ipv4, pkt.Udp, gopacket.Payload(pkt.Payload))
	}
	if pkt.Tcp != nil {
		pkt.Tcp.SetNetworkLayerForChecksum(pkt.Ipv4)
		gopacket.SerializeLayers(buf, opts, pkt.Ethernet, pkt.Ipv4, pkt.Tcp, gopacket.Payload(pkt.Payload))
	}
	if pkt.Icmpv4 != nil {
		gopacket.SerializeLayers(buf, opts, pkt.Ethernet, pkt.Ipv4, pkt.Icmpv4, gopacket.Payload(pkt.Payload))
	}
	err = s.io.WritePacketData(buf.Bytes())
	return
}

func (s *Server) sendPackets() {
	var err error
	defer close(s.TxChan)

	for {
		select {
		case pkt := <-s.TxChan:
			buf := gopacket.NewSerializeBuffer()
			opts := gopacket.SerializeOptions{
				FixLengths:       true,
				ComputeChecksums: true,
			}
			if pkt.Udp != nil {
				pkt.Udp.SetNetworkLayerForChecksum(pkt.Ipv4)
				gopacket.SerializeLayers(buf, opts, pkt.Ethernet, pkt.Ipv4, pkt.Udp, gopacket.Payload(pkt.Payload))
			}
			if pkt.Tcp != nil {
				pkt.Tcp.SetNetworkLayerForChecksum(pkt.Ipv4)
				gopacket.SerializeLayers(buf, opts, pkt.Ethernet, pkt.Ipv4, pkt.Tcp, gopacket.Payload(pkt.Payload))
			}
			if pkt.Icmpv4 != nil {
				gopacket.SerializeLayers(buf, opts, pkt.Ethernet, pkt.Ipv4, pkt.Icmpv4, gopacket.Payload(pkt.Payload))
			}
			err = s.io.WritePacketData(buf.Bytes())
			if err != nil {
				log.Fatalf("write packet data err : %+v", err)
			}

			s.stats.Tx.Packets++
		}
	}
	return
}

func (s *Server) readPackets() {
	packetSource := s.io.PacketSource()
	defer close(s.RxChan)
	for {

		packet, err := packetSource.NextPacket()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Printf("[ERROR] readPackets err: %s", err)
			continue
		}
		s.stats.Rx.Packets++

		s.RxChan <- packet
		if s.isStopped {
			break
		}
	}
}

func (s *Server) Shutdown() {
	s.io.Close()
	s.isStopped = true
}

func (s *Server) Stats() stats.Directional {
	return s.stats
}

func (s *Server) signalWorker() {
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
