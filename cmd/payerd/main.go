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

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func _main(ctxCli *cli.Context) error {
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
	if err := app.Run(os.Args); err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}
}

var app = cli.NewApp()

var cliFlags = []cli.Flag{
	ConfigFileFlag,
	VerbosityFlag,
	KeystorePathFlag,
	ConfigPath,
}

var generateFlags = []cli.Flag{
	PathFlag,
}

var accountCommand = cli.Command{
	Name:        "accounts",
	Usage:       "manage payer keystore",
	Description: "The accounts command is used to manage the payer keystore.\n",
	Subcommands: []*cli.Command{
		{
			Action: wrapHandler(handleGenerateSubCmd),
			Name:   "gensub",
			Usage:  "generate subsrate keystore",
			Flags:  generateFlags,
			Description: "The generate subcommand is used to generate the substrate keystore.\n" +
				"\tkeystore path should be given.",
		},
	},
}

// init initializes CLI
func init() {
	app.Action = _main
	app.Copyright = "Copyright 2021 Stafi Protocol Authors"
	app.Name = "payer"
	app.Usage = "payer"
	app.Authors = []*cli.Author{{Name: "Stafi Protocol 2021"}}
	app.Version = "0.0.1"
	app.EnableBashCompletion = true
	app.Commands = []*cli.Command{
		&accountCommand,
	}

	app.Flags = append(app.Flags, cliFlags...)
}
