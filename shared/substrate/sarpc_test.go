package substrate_test

import (
	"fee-station/shared/substrate"
	"testing"
	"time"

	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

func TestGetSomeEvents(t *testing.T) {
	// endpoint := "wss://polkadot-rpc3.stafi.io"
	endpoint := "wss://kusama-rpc.polkadot.io"
	// endpoint ="wss://mainnet-rpc.stafi.io"
	gc, err := substrate.NewGsrpcClient(endpoint, substrate.AddressTypeAccountId, nil)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := substrate.NewSarpcClient(substrate.ChainTypePolkadot, endpoint, "/Users/tpkeeper/gowork/stafi/fee-station/network/kusama.json")
	if err != nil {
		t.Fatal(err)
	}
	finalNumber, err := gc.GetFinalizedBlockNumber()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("finalNumber:", finalNumber)

	for i := finalNumber; i > 0; i++ {
		t.Log("now deal number", i)
		var hash types.Hash
		for {

			hashStr, err := sc.GetBlockHash(i)
			if err != nil {
				t.Log(err)
				time.Sleep(time.Second * 1)
				continue
			}

			hash, err = types.NewHashFromHexString(hashStr)
			if err != nil {
				t.Log(err)
				time.Sleep(time.Second * 1)
				continue
			}
			break
		}

		number, err := gc.GetBlockNumber(hash)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("number:", number)

		extrinsics, err := sc.GetExtrinsics(hash.Hex())
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range extrinsics {
			t.Log(n.Address, n.CallModuleName, n.CallName, n.Params)
		}

		event, err := sc.GetChainEvents(hash.Hex())
		if err != nil {
			t.Fatal(err)
		}

		for _, e := range event {
			t.Log(e.EventId, e.ModuleId, e.Params)
		}
	}

}

func TestGetEvent(t *testing.T) {
	// endpoint := "wss://polkadot-rpc3.stafi.io"
	endpoint := "wss://kusama-rpc.polkadot.io"
	// endpoint ="wss://mainnet-rpc.stafi.io"
	gc, err := substrate.NewGsrpcClient(endpoint, substrate.AddressTypeAccountId, nil)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := substrate.NewSarpcClient(substrate.ChainTypePolkadot, endpoint, "/Users/tpkeeper/gowork/stafi/fee-station/network/kusama.json")
	if err != nil {
		t.Fatal(err)
	}
	finalNumber, err := gc.GetFinalizedBlockNumber()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("finalNumber:", finalNumber)
	i := uint64(9674430)
	t.Log("now deal number", i)
	hashStr, err := sc.GetBlockHash(i)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := types.NewHashFromHexString(hashStr)
	if err != nil {
		t.Fatal(hash)
	}
	number, err := gc.GetBlockNumber(hash)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("number:", number)

	extrinsics, err := sc.GetExtrinsics(hash.Hex())
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range extrinsics {
		t.Log(n.Address, n.CallModuleName, n.CallName, n.Params)
	}

	event, err := sc.GetChainEvents(hash.Hex())
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range event {
		t.Log(e.EventId, e.ModuleId, e.Params)
	}

}


func TestTransferParam(t *testing.T) {
	endpoint := "wss://polkadot-rpc3.stafi.io"
	// endpoint := "wss://kusama-rpc.polkadot.io"
	// endpoint ="wss://mainnet-rpc.stafi.io"
	gc, err := substrate.NewGsrpcClient(endpoint, substrate.AddressTypeAccountId, nil)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := substrate.NewSarpcClient(substrate.ChainTypePolkadot, endpoint, "/Users/tpkeeper/gowork/stafi/fee-station/network/polkadot.json")
	if err != nil {
		t.Fatal(err)
	}

	hashStr:="0x268f7c1d4b425f5ae671b0f41514c6e25cf959c44f76e70c48d3b217499bf572"
	hash, err := types.NewHashFromHexString(hashStr)
	if err != nil {
		t.Fatal(hash)
	}
	number, err := gc.GetBlockNumber(hash)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("number:", number)

	extrinsics, err := sc.GetExtrinsics(hash.Hex())
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range extrinsics {

		t.Log(n.Address, n.CallModuleName, n.CallName, n.Params)
	}

	event, err := sc.GetChainEvents(hash.Hex())
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range event {
		t.Log(e.EventId, e.ModuleId, e.Params)
	}

}
