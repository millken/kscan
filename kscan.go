package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/google/gopacket/examples/util"
	"github.com/hashicorp/logutils"
	"github.com/millken/kscan/config"
	"github.com/millken/kscan/server"
)

var VERSION string = "2.0.0"

func main() {
	var err error
	var (
		configPath = flag.String("c", "config.toml", "config path")
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
	cf, err := config.Load(*configPath)
	if err != nil {
		log.Printf("[ERROR] %s", err.Error())
		return
	}
	filter_writer := os.Stderr
	if cf.Log.File != "" {
		filter_writer, err = os.Create(cf.Log.File)
		if err != nil {
			log.Printf("[ERROR] %s", err)
		}
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"FINE", "DEBUG", "TRACE", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel(cf.Log.Level),
		Writer:   filter_writer,
	}
	log.SetOutput(filter)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("[INFO] Loading config : %s, version: %s", *configPath, VERSION)

	log.Printf("[DEBUG] config= %v , level=%s", cf, cf.Log.Level)

	s := server.New(cf)
	if err = s.Start(); err != nil {
		log.Printf("[ERROR] :%s", err)
	}
}
