package drivers

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pfring"
)

func init() {
	DriverRegister("pfring", NewPfringHandle)
}

type PfringHandle struct {
	handle *pfring.Ring
}

func NewPfringHandle(options *DriverOptions) (PacketDataSourceCloser, error) {
	var flags pfring.Flag
	if options.Device == "any" {
		return nil, fmt.Errorf("Pfring sniffing doesn't support 'any' as interface")
	}
	//flags = pfring.FlagPromisc
	pfringWireHandle, err := pfring.NewRing(options.Device, uint32(options.Snaplen), flags)
	pfringHandle := PfringHandle{
		handle: pfringWireHandle,
	}
	if err != nil {
		return &pfringHandle, err
	}
	if err = pfringHandle.handle.Enable(); err != nil {
		return &pfringHandle, err

	}
	err = pfringHandle.handle.SetBPFFilter(options.Filter)
	return &pfringHandle, err
}

func (p *PfringHandle) PacketSource() *gopacket.PacketSource {
	return nil
}

func (p *PfringHandle) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	return p.handle.ReadPacketData()
}

func (p *PfringHandle) WritePacketData(data []byte) (err error) {
	return p.handle.WritePacketData(data)
}

func (p *PfringHandle) Close() {
	if p.handle != nil {
		p.handle.Close()
		p.handle = nil
	}
}
