// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	TaskTicker    int64 //seconds task interval
	EthApi        string
	REthStatApi   string
	LedgerApi     string
	MaxGasPrice   int64 //Gwei
	ChainId       int64
	DropContract  string
	DropTime      int64 //seconds from utc+8 00:00
	SyncStartDate string
	From          string
	KeystorePath  string
	DropIsOpen    bool

	ListenAddr  string
	Mode        string //release debug test
	Db          Db
	LogFilePath string
}

type Db struct {
	Host string
	Port string
	User string
	Pwd  string
	Name string
}

func Load(defaultCfgFile string) (*Config, error) {
	configFilePath := flag.String("C", defaultCfgFile, "Config file path")
	flag.Parse()

	var cfg = Config{}
	if err := loadSysConfig(*configFilePath, &cfg); err != nil {
		return nil, err
	}
	if cfg.LogFilePath == "" {
		cfg.LogFilePath = "./log_data"
	}
	return &cfg, nil
}

func loadSysConfig(path string, config *Config) error {
	_, err := os.Open(path)
	if err != nil {
		return err
	}
	if _, err := toml.DecodeFile(path, config); err != nil {
		return err
	}
	fmt.Println("load sysConfig success")
	return nil
}
