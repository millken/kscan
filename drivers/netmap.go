// +build freebsd

package drivers

import "github.com/google/gopacket"

func init() {
	DriverRegister("netmap", NewNetmapHandle)
}

type NetmapHandle struct {
	handle *gonetmap.Netmap
}

func NewNetmapHandle(options *DriverOptions) (PacketDataSourceCloser, error) {
	wireHandle, err := gonetmap.OpenNetmap(options.Device)
	netmapHandle := NetmapHandle{
		handle: WireHandle,
	}

	return &netmapHandle, err
}

func (h *NetmapHandle) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
}

func (h *NetmapHandle) WritePacketData(data []byte) (err error) {
	return h.handle.Inject(data)
}

func (h *NetmapHandle) close() {
	if h.handle != nil {
		h.handle.Close()
		h.handle = nil
	}
}
