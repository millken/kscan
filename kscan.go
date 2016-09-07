package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/gopacket/examples/util"
	"github.com/hashicorp/logutils"
	"github.com/millken/kscan/config"
	"github.com/millken/kscan/program"
	"github.com/millken/kscan/server"
)

var VERSION string = "2.0.0"

func main() {
	var err error
	var (
		configPath  = flag.String("c", "config.toml", "config path")
		programName = flag.String("p", "sample", "program select")
		//help       = flag.String("h", "", "usage")
	)
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic ->>>> %s", err)
		}
	}()
	defer util.Run()()

	if os.Geteuid() != 0 {
		log.Printf("requires root!")
		return
	}
	cf, mcf, err := config.Load(*configPath)
	if err != nil {
		log.Printf("[ERROR] %s", err.Error())
		return
	}
	filter_writer := os.Stderr
	if cf.LogFile != "" {
		filter_writer, err = os.Create(cf.LogFile)
		if err != nil {
			log.Printf("[ERROR] %s", err)
		}
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"FINE", "DEBUG", "TRACE", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(cf.LogLevel),
		Writer:   filter_writer,
	}
	log.SetOutput(filter)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("[INFO] Loading config : %s, version: %s", *configPath, VERSION)

	log.Printf("[DEBUG] master config= %v , mode config = %v,level=%s", cf, mcf, cf.LogLevel)

	s := server.New(cf, mcf)
	if err = s.Start(); err != nil {
		log.Printf("[ERROR] :%s", err)
	}
	if err = program.Start(*programName, s); err != nil {
		log.Printf("[ERROR] :%s", err)
	}
	//waiting 3 second for receive data
	time.Sleep(3 * time.Second)
}
