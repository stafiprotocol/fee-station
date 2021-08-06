package rpc

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptoTypes "github.com/cosmos/cosmos-sdk/crypto/types"
	xAuthTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	rpcHttp "github.com/tendermint/tendermint/rpc/client/http"
	"os"
)

//cosmos client
type Client struct {
	clientCtx     client.Context
	rpcClient     rpcClient.Client
	denom         string
	endPoint      string
	accountNumber uint64
}

func NewClient(denom, endPoint string) (*Client, error) {
	encodingConfig := MakeEncodingConfig()
	rpcClient, err := rpcHttp.New(endPoint, "/websocket")
	if err != nil {
		return nil, err
	}

	initClientCtx := client.Context{}.
		WithJSONMarshaler(encodingConfig.Marshaler).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(xAuthTypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithClient(rpcClient).
		WithSkipConfirmation(true) //skip password confirm

	client := &Client{
		clientCtx: initClientCtx,
		rpcClient: rpcClient,
		endPoint:  endPoint,
	}

	account, err := client.GetAccount()
	if err != nil {
		return nil, err
	}
	client.accountNumber = account.GetAccountNumber()

	client.setDenom(denom)
	if err != nil {
		return nil, err
	}
	return client, nil
}

//update clientCtx.FromName and clientCtx.FromAddress
func (c *Client) SetFromName(fromName string) error {
	info, err := c.clientCtx.Keyring.Key(fromName)
	if err != nil {
		return fmt.Errorf("keyring get address from fromKeyname err: %s", err)
	}

	c.clientCtx = c.clientCtx.WithFromName(fromName).WithFromAddress(info.GetAddress())

	account, err := c.GetAccount()
	if err != nil {
		return err
	}
	c.accountNumber = account.GetAccountNumber()
	return nil
}

func (c *Client) GetFromName() string {
	return c.clientCtx.FromName
}

func (c *Client) setDenom(denom string) {
	c.denom = denom
}

func (c *Client) GetDenom() string {
	return c.denom
}

func (c *Client) GetTxConfig() client.TxConfig {
	return c.clientCtx.TxConfig
}

func (c *Client) GetLegacyAmino() *codec.LegacyAmino {
	return c.clientCtx.LegacyAmino
}

func (c *Client) Sign(fromName string, toBeSigned []byte) ([]byte, cryptoTypes.PubKey, error) {
	return c.clientCtx.Keyring.Sign(fromName, toBeSigned)
}
