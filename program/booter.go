package program

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
)

type Booter interface {
	Init(config map[string]toml.Primitive) error
}

var booters = make(map[string]func() interface{})

func RegisterBooter(name string, booter func() interface{}) {
	if booter == nil {
		log.Fatalln("booter: Register booter is nil")
	}

	if _, ok := booters[name]; ok {
		log.Fatalln("booter: Register called twice for booter " + name)
	}
	log.Printf("Register %s, %v", name, booter)

	booters[name] = booter
}

func Start(name string, mcf map[string]toml.Primitive) (err error) {
	if b, ok := booters[name]; ok {

		return b().(Booter).Init(mcf)
	}
	return fmt.Errorf("program %s not exist", name)
}
