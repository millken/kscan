package sample

const NAME = "sample"

type Config struct {
	SrcMac string `toml:"src_mac"`
	DstMac string `toml:"dst_mac"`
	DstIp  string `toml:"dst_ip"`
}
