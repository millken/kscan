package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

type MasterConf struct {
	Iface, Driver string
	Filter        string
	WorkerNum     int    `toml:"worker_num"`
	LogFile       string `toml:"log_file"`
	LogLevel      string `toml:"log_level"`
}

type ProgramConf map[string]toml.Primitive

func Load(configPath string) (masterConfig *MasterConf, programConfig ProgramConf, err error) {

	var configFile ProgramConf
	p, err := os.Open(configPath)
	if err != nil {
		return nil, nil, fmt.Errorf("Error opening config file: %s", err)
	}
	contents, err := ioutil.ReadAll(p)
	if err != nil {
		return nil, nil, fmt.Errorf("Error reading config file: %s", err)
	}
	if _, err = toml.Decode(string(contents), &configFile); err != nil {
		return nil, nil, fmt.Errorf("Error decoding config file: %s", err)
	}
	parsed_config, ok := configFile["master"]
	if ok {
		if err = toml.PrimitiveDecode(parsed_config, &masterConfig); err != nil {
			err = fmt.Errorf("Can't unmarshal master config: %s", err)
		}
	}
	programConfig = configFile
	delete(programConfig, "master")
	return
}
