package drivers

import "github.com/google/gopacket"

type DriverOptions struct {
	Device  string
	Snaplen int32
	Filter  string
}

type PacketDataSourceCloser interface {
	ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error)
	WritePacketData(data []byte) (err error)
	PacketSource() *gopacket.PacketSource
	Close()
}
