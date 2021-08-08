// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/log"
	"fee-station/pkg/utils"
	"fee-station/task/payer"
	"fmt"
	"github.com/stafiprotocol/chainbridge/utils/crypto/sr25519"
	"github.com/stafiprotocol/chainbridge/utils/keystore"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

func _main() error {
	cfg, err := config.Load("conf_payer.toml")
	if err != nil {
		fmt.Printf("loadConfig err: %s", err)
		return err
	}
	log.InitLogFile(cfg.LogFilePath + "/payer")
	logrus.Infof("config info:%+v ", cfg)

	kp, err := keystore.KeypairFromAddress(cfg.PayerAccount, keystore.SubChain, cfg.KeystorePath, false)
	if err != nil {
		return fmt.Errorf("keypairFromAddress err: %s", err)
	}
	krp := kp.(*sr25519.Keypair).AsKeyringPair()

	//init db
	db, err := db.NewDB(&db.Config{
		Host:   cfg.Db.Host,
		Port:   cfg.Db.Port,
		User:   cfg.Db.User,
		Pass:   cfg.Db.Pwd,
		DBName: cfg.Db.Name,
		Mode:   cfg.Mode})
	if err != nil {
		logrus.Errorf("db err: %s", err)
		return err
	}
	logrus.Infof("db connect success")

	//interrupt signal
	ctx := utils.ShutdownListener()
	defer func() {
		sqlDb, err := db.DB.DB()
		if err != nil {
			logrus.Errorf("db.DB() err: %s", err)
			return
		}
		logrus.Infof("shutting down the db ...")
		sqlDb.Close()
	}()
	t := task.NewTask(cfg, db, krp)
	err = t.Start()
	if err != nil {
		logrus.Errorf("task start err: %s", err)
		return err
	}
	defer func() {
		logrus.Infof("shutting down task ...")
		t.Stop()
	}()

	<-ctx.Done()
	return nil
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	debug.SetGCPercent(40)
	err := _main()
	if err != nil {
		os.Exit(1)
	}
}
