package server

import (
	"github.com/google/gopacket"
	"github.com/millken/kscan/layer"
)

func packetHandler(i int, in <-chan gopacket.Packet, out chan layer.Packet) {
}
