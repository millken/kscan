package drivers

var Drivers = map[string]func(*DriverOptions) (PacketDataSourceCloser, error){}

// Register makes a ethernet sniffer driver available by the provided name.
// If Register is called twice with the same name or if driver is nil, it panics.
func DriverRegister(name string, packetDataSourceCloserFactory func(*DriverOptions) (PacketDataSourceCloser, error)) {
	if packetDataSourceCloserFactory == nil {
		panic(" packetDataSourceCloserFactory is nil")
	}
	if _, dup := Drivers[name]; dup {
		panic(" Register called twice for ethernet sniffer " + name)
	}
	Drivers[name] = packetDataSourceCloserFactory
}
