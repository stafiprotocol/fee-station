package substrate_test

import (
	"fee-station/shared/substrate"
	"testing"
)

func TestGetEvent(t *testing.T) {
	endpoint := "wss://polkadot-rpc3.stafi.io"
	// endpoint ="wss://mainnet-rpc.stafi.io"
	sc, err := substrate.NewSarpcClient(substrate.ChainTypePolkadot, endpoint, "/Users/tpkeeper/gowork/stafi/fee-station/network/polkadot.json")
	if err != nil {
		t.Fatal(err)
	}
	// need, err := sc.GetExtrinsics("0x4bb084f0914628b2688acd82cd161c2c48dbfd65017f8469357931f3bc8a07b7")
	for {
		// need, err := sc.GetExtrinsics("0x9220f285c97971b7b2b3ac6ee614cfb2760f383d0dd3abc0d2f68ec56234f829")
		need, err := sc.GetExtrinsics("0x5487b78630c24312b56953fd493e0c2900e85cc7a91fd61c8c214e9f7fedc66c")
		if err != nil {
			t.Log(err)
		}
		for _, n := range need {
			t.Log(n.Address, n.CallModuleName, n.CallName, n.Params)
		}
	}

}
