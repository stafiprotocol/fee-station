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
	TaskTicker  int64 //seconds task interval
	LogFilePath string
	//checker
	DotTypesPath  string
	KsmTypesPath  string
	CoinMarketApi string
	CoinGeckoApi  string
	Endpoint      Endpoint
	// payer
	KeystorePath string
	PayerAccount string
	FisEndpoint  string
	//station
	ListenAddr   string
	AtomDenom    string
	SwapRate     string //decimals 6
	SwapMaxLimit string //decimals 12
	SwapMinLimit string //decimals 12
	Mode         string //release debug test
	PoolAddress  PoolAddress
	//syncer
	EtherScanApiKey string
	SubScanApiKey   string
	RecoverInterval int64
	StartTimestamp  int64
	SyncTxEndpoint  SyncTxEndpoint
	//common
	Db Db
}

type PoolAddress struct {
	Eth  string
	Atom string
	Dot  string
	Ksm  string
}

type Endpoint struct {
	Eth  string
	Atom string
	Dot  string
	Ksm  string
}

type SyncTxEndpoint struct {
	Eth  string
	Atom string
	Dot  string
	Ksm  string
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
