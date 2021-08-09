// Copyright 2020 Stafi Protocol
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/ChainSafe/log15"
	"github.com/stafiprotocol/chainbridge/utils/crypto"
	"github.com/stafiprotocol/chainbridge/utils/crypto/secp256k1"
	"github.com/stafiprotocol/chainbridge/utils/crypto/sr25519"
	"github.com/stafiprotocol/chainbridge/utils/keystore"
	"github.com/urfave/cli/v2"
)

const DefaultKeystorePath = "./keys"

var (
	ConfigFileFlag = &cli.StringFlag{
		Name:  "config",
		Usage: "TOML configuration file",
	}

	VerbosityFlag = &cli.StringFlag{
		Name:  "verbosity",
		Usage: "Supports levels crit (silent) to trce (trace)",
		Value: log.LvlInfo.String(),
	}

	KeystorePathFlag = &cli.StringFlag{
		Name:  "keystore",
		Usage: "Path to keystore directory",
		Value: DefaultKeystorePath,
	}

	BlockstorePathFlag = &cli.StringFlag{
		Name:  "blockstore",
		Usage: "Specify path for blockstore",
		Value: "", // Empty will use home dir
	}

	FreshStartFlag = &cli.BoolFlag{
		Name:  "fresh",
		Usage: "Disables loading from blockstore at start. Opts will still be used if specified.",
	}

	LatestBlockFlag = &cli.BoolFlag{
		Name:  "latest",
		Usage: "Overrides blockstore and start block, starts from latest block",
	}
	ConfigPath = &cli.StringFlag{
		Name:  "C",
		Usage: "Path to configfile",
		Value: "./conf_payer.toml",
	}
)

// Metrics flags
var (
	MetricsFlag = &cli.BoolFlag{
		Name:  "metrics",
		Usage: "Enables metric server",
	}

	MetricsPort = &cli.IntFlag{
		Name:  "metricsPort",
		Usage: "Port to serve metrics on",
		Value: 8001,
	}
)

// Generate subcommand flags
var (
	PathFlag = &cli.StringFlag{
		Name:  "keypath",
		Usage: "Dir to keep key file.",
	}
)

//dataHandler is a struct which wraps any extra data our CMD functions need that cannot be passed through parameters
type dataHandler struct {
	datadir string
}

// wrapHandler takes in a Cmd function (all declared below) and wraps
// it in the correct signature for the Cli Commands
func wrapHandler(hdl func(*cli.Context, *dataHandler) error) cli.ActionFunc {

	return func(ctx *cli.Context) error {

		datadir, err := getDataDir(ctx)
		if err != nil {
			return fmt.Errorf("failed to access the datadir: %s", err)
		}

		return hdl(ctx, &dataHandler{datadir: datadir})
	}
}

func handleGenerateSubCmd(ctx *cli.Context, dHandler *dataHandler) error {
	log.Info("Generating substrate keyfile by rawseed...")
	path := ctx.String(PathFlag.Name)
	return generateKeyFileByRawseed(path)
}

func handleGenerateEthCmd(ctx *cli.Context, dHandler *dataHandler) error {
	log.Info("Generating ethereum keyfile by private key...")
	path := ctx.String(PathFlag.Name)
	return generateKeyFileByPrivateKey(path)
}

// getDataDir obtains the path to the keystore and returns it as a string
func getDataDir(ctx *cli.Context) (string, error) {
	// key directory is datadir/keystore/
	if dir := ctx.String(KeystorePathFlag.Name); dir != "" {
		datadir, err := filepath.Abs(dir)
		if err != nil {
			return "", err
		}
		log.Trace(fmt.Sprintf("Using keystore dir: %s", datadir))
		return datadir, nil
	}
	return "", fmt.Errorf("datadir flag not supplied")
}

// keypath example: /Homepath/chainbridge/keys
func generateKeyFileByRawseed(keypath string) error {
	key := keystore.GetPassword("Enter mnemonic/rawseed:")
	kp, err := sr25519.NewKeypairFromSeed(string(key), "stafi")
	if err != nil {
		return err
	}

	fp, err := filepath.Abs(keypath + "/" + kp.Address() + ".key")
	if err != nil {
		return fmt.Errorf("invalid filepath: %s", err)
	}

	file, err := os.OpenFile(filepath.Clean(fp), os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Error("generate keypair: could not close keystore file")
		}
	}()

	password := keystore.GetPassword("password for key:")
	err = keystore.EncryptAndWriteToFile(file, kp, password)
	if err != nil {
		return fmt.Errorf("could not write key to file: %s", err)
	}

	log.Info("key generated", "address", kp.Address(), "type", "sub", "file", fp)
	return nil
}

func generateKeyFileByPrivateKey(keypath string) error {
	var kp crypto.Keypair
	var err error

	key := keystore.GetPassword("Enter private key:")
	skey := string(key)

	if skey[0:2] == "0x" {
		kp, err = secp256k1.NewKeypairFromString(skey[2:])
	} else {
		kp, err = secp256k1.NewKeypairFromString(skey)
	}
	if err != nil {
		return fmt.Errorf("could not generate secp256k1 keypair from given string: %s", err)
	}

	fp, err := filepath.Abs(keypath + "/" + kp.Address() + ".key")
	if err != nil {
		return fmt.Errorf("invalid filepath: %s", err)
	}

	file, err := os.OpenFile(filepath.Clean(fp), os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			log.Error("generate keypair: could not close keystore file")
		}
	}()

	password := keystore.GetPassword("password for key:")
	err = keystore.EncryptAndWriteToFile(file, kp, password)
	if err != nil {
		return fmt.Errorf("could not write key to file: %s", err)
	}

	log.Info("key generated", "address", kp.Address(), "type", "eth", "file", fp)
	return nil
}
