package rpc_test

import (
	"bytes"
	"encoding/hex"
	"sort"
	"sync"
	"testing"
	"time"

	"fee-station/shared/cosmos/rpc"
	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/cosmos/cosmos-sdk/types"
	// xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

var client *rpc.Client

//eda331e37bf66b2393c4c271e384dfaa2bfcdd35
var addrMultiSig1, _ = types.AccAddressFromBech32("cosmos1ak3nrcmm7e4j8y7ycfc78pxl4g4lehf43vw6wu")
var addrReceive, _ = types.AccAddressFromBech32("cosmos1cgs647rewxyzh5wu4e606kk7qyuj5f8hk20rgf")
var addrValidator, _ = types.ValAddressFromBech32("cosmosvaloper1y6zkfcvwkpqz89z7rwu9kcdm4kc7uc4e5y5a2r")
var addrKey1, _ = types.AccAddressFromBech32("cosmos1a8mg9rj4nklhmwkf5vva8dvtgx4ucd9yjasret")
var addrValidatorTestnet2, _ = types.ValAddressFromBech32("cosmosvaloper19xczxvvdg8h67sk3cccrvxlj0ruyw3360rctfa")

var addrValidatorTestnet, _ = types.ValAddressFromBech32("cosmosvaloper17tpddyr578avyn95xngkjl8nl2l2tf6auh8kpc")
var addrValidatorTestnetStation, _ = types.ValAddressFromBech32("cosmosvaloper1x5wgh6vwye60wv3dtshs9dmqggwfx2ldk5cvqu")
var addrValidatorTestnetAteam, _ = types.ValAddressFromBech32("cosmosvaloper105gvcjgs6s4j5ws9srckx0drt4x8cwgywplh7p")

var adrValidatorTestnetTecos, _ = types.ValAddressFromBech32("cosmosvaloper1p7e37nztj62mmra8xhgqde7sql3llhhu6hvcx8")
var adrValidatorEverStake, _ = types.ValAddressFromBech32("cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3")
var adrValidatorForbole, _ = types.ValAddressFromBech32("cosmosvaloper1w96rrh9sx0h7n7qak00l90un0kx5wala2prmxt")

func TestGetAddrHex(t *testing.T) {
	t.Log("cosmosvaloper17tpddyr578avyn95xngkjl8nl2l2tf6auh8kpc", hexutil.Encode(addrValidatorTestnet.Bytes()))
	t.Log("cosmosvaloper1x5wgh6vwye60wv3dtshs9dmqggwfx2ldk5cvqu", hexutil.Encode(addrValidatorTestnetStation.Bytes()))
	t.Log("cosmosvaloper105gvcjgs6s4j5ws9srckx0drt4x8cwgywplh7p", hexutil.Encode(addrValidatorTestnetAteam.Bytes()))

	t.Log("cosmosvaloper1p7e37nztj62mmra8xhgqde7sql3llhhu6hvcx8", hexutil.Encode(adrValidatorTestnetTecos.Bytes()))
	t.Log("cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3", hexutil.Encode(adrValidatorEverStake.Bytes()))
	t.Log("cosmosvaloper1w96rrh9sx0h7n7qak00l90un0kx5wala2prmxt", hexutil.Encode(adrValidatorForbole.Bytes()))
	//client_test.go:36: cosmosvaloper17tpddyr578avyn95xngkjl8nl2l2tf6auh8kpc 0xf2c2d69074f1fac24cb434d1697cf3fabea5a75d
	//client_test.go:38: cosmosvaloper1x5wgh6vwye60wv3dtshs9dmqggwfx2ldk5cvqu 0x351c8be98e2674f7322d5c2f02b760421c932bed
	//client_test.go:37: cosmosvaloper105gvcjgs6s4j5ws9srckx0drt4x8cwgywplh7p 0x7d10cc4910d42b2a3a0580f1633da35d4c7c3904

	//client_test.go:40: cosmosvaloper1p7e37nztj62mmra8xhgqde7sql3llhhu6hvcx8 0x0fb31f4c4b9695bd8fa735d006e7d007e3ffdefc
	//client_test.go:41: cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3 0x5a7f68bf60a3100937e42aad6bdc111f72c55e62
	//client_test.go:42: cosmosvaloper1w96rrh9sx0h7n7qak00l90un0kx5wala2prmxt 0x717431dcb033efe9f81db3dff2bf937d8d4777fd
}

func init() {

	// client, err = rpc.NewClient(key, "stargate-final", "recipient", "0.04umuon", "umuon", "https://testcosmosrpc.wetez.io:443")
	var err error
	// client, err = rpc.NewClient("umuon", "http://127.0.0.1:26657")
	client, err = rpc.NewClient("stake", "https://cosmos-rpc1.stafi.io:443")
	if err != nil {
		panic(err)
	}
}

func TestClient_QueryTxByHash(t *testing.T) {
	res, err := client.QueryTxByHash("6C017062FD3F48F13B640E5FEDD59EB050B148E67EF12EC0A511442D32BD4C88")
	t.Log(err)
	assert.NoError(t, err)
	for _, msg := range res.GetTx().GetMsgs() {

		t.Log(msg.String())
		t.Log(msg.Type())
		t.Log(msg.Route())
	}
}

func TestGetPubKey(t *testing.T) {
	test, err := types.AccAddressFromBech32("cosmos12zhwz792d8zpxj3wmz05c7k9meea6q0xvf5y79")
	assert.NoError(t, err)
	account, err := client.QueryAccount(test)
	assert.NoError(t, err)
	t.Log(hex.EncodeToString(account.GetPubKey().Bytes()))

	// res, err := client.QueryTxByHash("327DA2048B6D66BCB27C0F1A6D1E407D88FE719B95A30D108B5906FD6934F7B1")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// msgs := res.GetTx().GetMsgs()
	// for i, _ := range msgs {
	// 	if msgs[i].Type() == xBankTypes.TypeMsgSend {
	// 		msg, _ := msgs[i].(*xBankTypes.MsgSend)
	// 		t.Log(msg.Amount.AmountOf("umuon").Uint64())
	// 	}
	// }

}

func TestGetEvents(t *testing.T) {
	events, err := client.GetEvents(
		[]string{"transfer.recipient='cosmos1jacw22mwkaswlml3ra7w7ue4fdak33kyx9fc8x'"},
		1, 10, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(events)
}

func TestClient_Sign(t *testing.T) {
	bts, err := hex.DecodeString("0E4F8F8FF7A3B67121711DA17FBE5AE8CB25DB272DDBF7DC0E02122947266604")
	assert.NoError(t, err)
	sigs, pubkey, err := client.Sign("recipient", bts)
	assert.NoError(t, err)
	t.Log(hex.EncodeToString(sigs))
	//4c6902bda88424923c62f95b3e3ead40769edab4ec794108d1c18994fac90d490087815823bd1a8af3d6a0271538cef4622b4b500a6253d2bd4c80d38e95aa6d
	t.Log(hex.EncodeToString(pubkey.Bytes()))
	//02e7710b4f7147c10ad90da06b69d2d6b8ff46786ef55a3f1e889c33de2bf0b416
}

func TestAddress(t *testing.T) {
	addrKey1, _ := types.AccAddressFromBech32("cosmos1a8mg9rj4nklhmwkf5vva8dvtgx4ucd9yjasret")
	addrKey2, _ := types.AccAddressFromBech32("cosmos1ztquzhpkve7szl99jkugq4l8jtpnhln76aetam")
	addrKey3, _ := types.AccAddressFromBech32("cosmos12zz2hm02sxe9f4pwt7y5q9wjhcu98vnuwmjz4x")
	addrKey4, _ := types.AccAddressFromBech32("cosmos12yprrdprzat35zhqxe2fcnn3u26gwlt6xcq0pj")
	pub, _ := types.GetPubKeyFromBech32(types.Bech32PubKeyTypeAccPub, "cosmospub1addwnpepqdj34s6y43njffwgpd83dr7smf03d9e5xzajt9lvqhen3avlfrnv7ya9n6t")

	t.Log(hex.EncodeToString(addrKey1.Bytes()))
	t.Log(hex.EncodeToString(addrKey2.Bytes()))
	t.Log(hex.EncodeToString(addrKey3.Bytes()))
	t.Log(hex.EncodeToString(addrKey4.Bytes()))
	t.Log(hex.EncodeToString(pub.Bytes())) //03651ac344ac6724a5c80b4f168fd0da5f16973430bb2597ec05f338f59f48e6cf
	//client_test.go:347: e9f6828e559dbf7dbac9a319d3b58b41abcc34a4
	//client_test.go:348: 12c1c15c36667d017ca595b88057e792c33bfe7e
	//client_test.go:349: 5084abedea81b254d42e5f894015d2be3853b27c
}

func TestClient_QueryDelegations(t *testing.T) {
	res, err := client.QueryDelegations(addrMultiSig1, 0)
	assert.NoError(t, err)
	t.Log(res.String())
}

func TestClient_QueryBalance(t *testing.T) {
	res, err := client.QueryBalance(addrMultiSig1, "umuon", 440000)
	assert.NoError(t, err)
	t.Log(res.Balance.Amount)
}

func TestClient_QueryDelegationTotalRewards(t *testing.T) {
	res, err := client.QueryDelegationTotalRewards(addrMultiSig1, 0)
	assert.NoError(t, err)
	t.Log(res.GetTotal().AmountOf(client.GetDenom()).TruncateInt())
}

func TestClient_GetSequence(t *testing.T) {
	seq, err := client.GetSequence(0, addrMultiSig1)
	assert.NoError(t, err)
	t.Log(seq)
	t.Log(hex.EncodeToString(addrValidatorTestnetAteam.Bytes()))
}

func TestMemo(t *testing.T) {
	res, err := client.QueryTxByHash("c7e3f7baf5a5f1d8cbc112080f32070dddd7cca5fe4272e06f8d42c17b25193f")
	assert.NoError(t, err)
	tx, err := client.GetTxConfig().TxDecoder()(res.Tx.GetValue())
	//tx, err := client.GetTxConfig().TxJSONDecoder()(res.Tx.Value)
	assert.NoError(t, err)
	memoTx, ok := tx.(types.TxWithMemo)
	assert.Equal(t, true, ok)
	t.Log(memoTx.GetMemo())
	hb, _ := hexutil.Decode("0xbebd0355ae360c8e6a7ed940a819838c66ca7b8f581f9c0e81dbb5faff346a30")
	//t.Log(string(hb))
	bonderAddr, _ := ss58.Encode(hb, ss58.StafiPrefix)
	t.Log(bonderAddr)
}

func TestMultiThread(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(50)

	for i := 0; i < 50; i++ {
		go func(i int) {
			t.Log(i)
			time.Sleep(5 * time.Second)
			height, err := client.GetAccount()
			if err != nil {
				t.Log("fail", i, err)
			} else {
				t.Log("success", i, height.GetSequence())
			}
			time.Sleep(15 * time.Second)
			height, err = client.GetAccount()
			if err != nil {
				t.Log("fail", i, err)
			} else {
				t.Log("success", i, height.GetSequence())
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestSort(t *testing.T) {
	a := []string{"cosmos1kuyde8vpt8c0ty4pxqgxw3makse7md80umthvg", "cosmos156kk2kqtwwwfps86g547swdlrc2cw6qctm6c8w", "cosmos1jkkhflu8qedqt4cyasd6tg70gjwx4jkhrse6rz"}
	t.Log(a)
	sort.SliceStable(a, func(i, j int) bool {
		return bytes.Compare([]byte(a[i]), []byte(a[j])) < 0
	})
	t.Log(a)
	rawTx := "7b22626f6479223a7b226d65737361676573223a5b7b224074797065223a222f636f736d6f732e62616e6b2e763162657461312e4d73674d756c746953656e64222c22696e70757473223a5b7b2261646472657373223a22636f736d6f7331776d6b39797334397a78676d78373770717337636a6e70616d6e6e7875737071753272383779222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a22313134373730227d5d7d5d2c226f757470757473223a5b7b2261646472657373223a22636f736d6f733135366b6b326b71747777776670733836673534377377646c7263326377367163746d36633877222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a2231393936227d5d7d2c7b2261646472657373223a22636f736d6f73316b7579646538767074386330747934707871677877336d616b7365376d643830756d74687667222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a223939383030227d5d7d2c7b2261646472657373223a22636f736d6f73316a6b6b68666c753871656471743463796173643674673730676a7778346a6b6872736536727a222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a223132393734227d5d7d5d7d5d2c226d656d6f223a22222c2274696d656f75745f686569676874223a2230222c22657874656e73696f6e5f6f7074696f6e73223a5b5d2c226e6f6e5f637269746963616c5f657874656e73696f6e5f6f7074696f6e73223a5b5d7d2c22617574685f696e666f223a7b227369676e65725f696e666f73223a5b5d2c22666565223a7b22616d6f756e74223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a2237353030227d5d2c226761735f6c696d6974223a2231353030303030222c227061796572223a22222c226772616e746572223a22227d7d2c227369676e617475726573223a5b5d7d"
	// rawTx:="7b22626f6479223a7b226d65737361676573223a5b7b224074797065223a222f636f736d6f732e62616e6b2e763162657461312e4d73674d756c746953656e64222c22696e70757473223a5b7b2261646472657373223a22636f736d6f7331776d6b39797334397a78676d78373770717337636a6e70616d6e6e7875737071753272383779222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a22313134373730227d5d7d5d2c226f757470757473223a5b7b2261646472657373223a22636f736d6f73316a6b6b68666c753871656471743463796173643674673730676a7778346a6b6872736536727a222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a223132393734227d5d7d2c7b2261646472657373223a22636f736d6f733135366b6b326b71747777776670733836673534377377646c7263326377367163746d36633877222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a2231393936227d5d7d2c7b2261646472657373223a22636f736d6f73316b7579646538767074386330747934707871677877336d616b7365376d643830756d74687667222c22636f696e73223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a223939383030227d5d7d5d7d5d2c226d656d6f223a22222c2274696d656f75745f686569676874223a2230222c22657874656e73696f6e5f6f7074696f6e73223a5b5d2c226e6f6e5f637269746963616c5f657874656e73696f6e5f6f7074696f6e73223a5b5d7d2c22617574685f696e666f223a7b227369676e65725f696e666f73223a5b5d2c22666565223a7b22616d6f756e74223a5b7b2264656e6f6d223a227561746f6d222c22616d6f756e74223a2237353030227d5d2c226761735f6c696d6974223a2231353030303030222c227061796572223a22222c226772616e746572223a22227d7d2c227369676e617475726573223a5b5d7d"
	txBts, err := hex.DecodeString(rawTx)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(txBts))
}
