package program

import (
	"fmt"
	"log"

	"github.com/millken/kscan/server"
)

type Booter interface {
	Init(s *server.Server) error
}

var booters = make(map[string]func() interface{})

func RegisterBooter(name string, booter func() interface{}) {
	if booter == nil {
		log.Fatalln("booter: Register booter is nil")
	}

	if _, ok := booters[name]; ok {
		log.Fatalln("booter: Register called twice for booter " + name)
	}

	booters[name] = booter
}

func Start(name string, s *server.Server) (err error) {
	if b, ok := booters[name]; ok {

		return b().(Booter).Init(s)
	}
	return fmt.Errorf("program %s not exist", name)
}
