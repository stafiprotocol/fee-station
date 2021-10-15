package substrate_test

import (
	"fee-station/shared/substrate"
	"testing"

	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

func TestGetEvent(t *testing.T) {
	endpoint := "wss://polkadot-rpc3.stafi.io"
	// endpoint ="wss://mainnet-rpc.stafi.io"
	gc, err := substrate.NewGsrpcClient(endpoint, substrate.AddressTypeAccountId, nil)
	if err != nil {
		t.Fatal(err)
	}
	finalNumber, err := gc.GetFinalizedBlockNumber()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("finalNumber:", finalNumber)

	hash, err := types.NewHashFromHexString("0x5487b78630c24312b56953fd493e0c2900e85cc7a91fd61c8c214e9f7fedc66c")
	if err != nil {
		t.Fatal(hash)
	}
	number, err := gc.GetBlockNumber(hash)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("number:", number)

	sc, err := substrate.NewSarpcClient(substrate.ChainTypePolkadot, endpoint, "/Users/tpkeeper/gowork/stafi/fee-station/network/polkadot.json")
	if err != nil {
		t.Fatal(err)
	}
	need, err := sc.GetExtrinsics("0x5487b78630c24312b56953fd493e0c2900e85cc7a91fd61c8c214e9f7fedc66c")
	if err != nil {
		t.Log(err)
	}
	for _, n := range need {
		t.Log(n.Address, n.CallModuleName, n.CallName, n.Params)
	}

}
