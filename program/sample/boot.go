package sample

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/millken/kscan/program"
)

type Boot struct {
}

var cf Config

func (t *Boot) Init(conf map[string]toml.Primitive) (err error) {
	log.Printf("[DEBUG] sample.conf = %v", conf)
	if _, ok := conf[NAME]; !ok {
		return fmt.Errorf("%s config is nil", NAME)
	}
	if err = toml.PrimitiveDecode(conf[NAME], &cf); err != nil {
		err = fmt.Errorf("Can't unmarshal config: %s", err)
	}
	log.Printf("[DEBUG] sample = %v", cf.SrcMac)
	return nil
}

func init() {
	program.RegisterBooter(NAME, func() interface{} {
		return new(Boot)
	})
}
