package task

import (
	"bytes"
	"encoding/hex"
	"errors"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	cosmosRpc "fee-station/shared/cosmos/rpc"
	"fmt"
	"time"

	"github.com/JFJun/go-substrate-crypto/ss58"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/shopspring/decimal"

	"github.com/ethereum/go-ethereum/common/hexutil"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/sirupsen/logrus"
)

// Frequency of polling for a new block
var (
	BlockRetryInterval = time.Second * 6
	BlockRetryLimit    = 50
	BlockConfirmNumber = int64(6)
)

func CheckAtomTx(db *db.WrapDb, denom, atomEndpoint string) error {
	swapInfoList, err := dao_station.GetSwapInfoListBySymbolState(db, utils.SymbolAtom, utils.SwapStateVerifySigs)
	if err != nil {
		return err
	}

	client, err := cosmosRpc.NewClient(denom, atomEndpoint)
	if err != nil {
		return err
	}
	for _, swapInfo := range swapInfoList {
		ok, err := TransferVerify(client, swapInfo)
		if err != nil {
			logrus.Errorf("atom TransferVerify failed: %s", err)
			continue
		}
		if ok {
			swapInfo.State = utils.SwapStateVerifyTxOk
		} else {
			swapInfo.State = utils.SwapStateVerifyTxFailed
		}
		err = dao_station.UpOrInSwapInfo(db, swapInfo)
		if err != nil {
			logrus.Warnf("dao_station.UpOrInSwapInfo err: %s", err)
			continue
		}
	}

	return nil
}

func TransferVerify(client *cosmosRpc.Client, swapInfo *dao_station.SwapInfo) (bool, error) {
	// hashStr := hex.EncodeToString(hexutil.MustDecode(swapInfo.Txhash))
	stafiAddressBytes, err := hexutil.Decode(swapInfo.StafiAddress)
	if err != nil {
		return false, err
	}
	blockHashBytes, err := hexutil.Decode(swapInfo.Blockhash)
	if err != nil {
		return false, err
	}
	txHashBytes, err := hexutil.Decode(swapInfo.Txhash)
	if err != nil {
		return false, err
	}
	pubkeyBytes, err := hexutil.Decode(swapInfo.Pubkey)
	if err != nil {
		return false, err
	}
	poolAddr, err := types.AccAddressFromBech32(swapInfo.PoolAddress)
	if err != nil {
		return false, err
	}

	inAmountDeci, err := decimal.NewFromString(swapInfo.InAmount)
	if err != nil {
		return false, err
	}

	hashStr := hex.EncodeToString(txHashBytes)
	//check tx hash
	txRes, err := GetTx(client, hashStr)
	if err != nil {
		return false, nil
	}

	if txRes.Empty() {
		return false, nil
	}

	if txRes.Code != 0 {
		return false, nil
	}

	//check block hash
	blockRes, err := GetBlock(client, txRes.Height)
	if err != nil {
		return false, err
	}
	if !bytes.Equal(blockRes.BlockID.Hash, blockHashBytes) {
		return false, nil
	}

	//check amount and pool
	amountIsMatch := false
	poolIsMatch := false
	var fromAddressStr string

	msgs := txRes.GetTx().GetMsgs()
	for i, _ := range msgs {
		if msgs[i].Type() == xBankTypes.TypeMsgSend {
			if sendMsg, ok := msgs[i].(*xBankTypes.MsgSend); ok {
				toAddr, err := types.AccAddressFromBech32(sendMsg.ToAddress)
				if err == nil {
					//amount and pool address must in one message
					if bytes.Equal(toAddr.Bytes(), poolAddr.Bytes()) &&
						sendMsg.Amount.AmountOf(client.GetDenom()).
							Equal(types.NewIntFromBigInt(inAmountDeci.BigInt())) {
						poolIsMatch = true
						amountIsMatch = true
						fromAddressStr = sendMsg.FromAddress
					}
				}

			}

		}
	}
	if !amountIsMatch {
		return false, nil
	}
	if !poolIsMatch {
		return false, nil
	}

	//check pubkey
	fromAddress, err := types.AccAddressFromBech32(fromAddressStr)
	if err != nil {
		return false, err
	}
	accountRes, err := client.QueryAccount(fromAddress)
	if err != nil {
		return false, err
	}

	if !bytes.Equal(accountRes.GetPubKey().Bytes(), pubkeyBytes) {
		return false, nil
	}
	//check memo
	var memoInTx string
	tx, err := client.GetTxConfig().TxDecoder()(txRes.Tx.GetValue())
	if err == nil {
		memoTx, ok := tx.(types.TxWithMemo)
		if ok {
			memoInTx = memoTx.GetMemo()
		}
	}

	bonderAddr, err := ss58.Encode(stafiAddressBytes, ss58.StafiPrefix)
	if err != nil {
		return false, nil // memo unmatch
	}

	if memoInTx != bonderAddr {
		return false, nil // memo unmatch
	}

	return true, nil
}

func GetTx(client *cosmosRpc.Client, txHash string) (*types.TxResponse, error) {
	var txRes *types.TxResponse
	var err error
	retryTx := 0
	for {
		if retryTx >= BlockRetryLimit {
			return nil, errors.New("QueryTxByHash reach retry limit")
		}
		txRes, err = client.QueryTxByHash(txHash)
		if err != nil {
			logrus.Warn(fmt.Sprintf("QueryTxByHash err: %s ,will retry queryTx after %f second", err, BlockRetryInterval.Seconds()))
			time.Sleep(BlockRetryInterval)
			retryTx++
			continue
		}
		currentHeight, err := client.GetCurrentBLockHeight()
		if err != nil {
			time.Sleep(BlockRetryInterval)
			retryTx++
			continue
		}
		if txRes.Height+BlockConfirmNumber > currentHeight {
			logrus.Warn(fmt.Sprintf("confirm number is smaller than %d ,will retry queryTx after %f second", BlockConfirmNumber, BlockRetryInterval.Seconds()))
			time.Sleep(BlockRetryInterval)
			retryTx++
			continue
		} else {
			break
		}

	}
	return txRes, nil
}

func GetBlock(client *cosmosRpc.Client, height int64) (*ctypes.ResultBlock, error) {
	var blockRes *ctypes.ResultBlock
	var err error
	retryTx := 0
	for {
		if retryTx >= BlockRetryLimit {
			return nil, errors.New("QueryBlock reach retry limit")
		}
		blockRes, err = client.QueryBlock(height)
		if err != nil {
			logrus.Warn(fmt.Sprintf("QueryBlock err: %s ,will retry queryBlock after %f second", err, BlockRetryInterval.Seconds()))
			time.Sleep(BlockRetryInterval)
			retryTx++
			continue
		}
		break
	}
	return blockRes, nil
}
