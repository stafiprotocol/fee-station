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

func CheckAtomTx(db *db.WrapDb, denom, atomEndpoint string) error {
	swapInfoList, err := dao_station.GetFeeStationSwapInfoListBySymbolState(db, utils.SymbolAtom, utils.SwapStateVerifySigs)
	if err != nil {
		return err
	}
	if len(swapInfoList) == 0 {
		return nil
	}

	retry := 0
	var client *cosmosRpc.Client
	for {
		if retry > BlockRetryLimit {
			return fmt.Errorf("cosmosRpc.NewClient reach retry limit")
		}

		client, err = cosmosRpc.NewClient(denom, atomEndpoint)
		if err != nil {
			logrus.Warnf("cosmosRpc newClient: %s", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		_, err = client.GetCurrentBLockHeight()
		if err != nil {
			logrus.Warnf("cosmosRpc newClient: %s", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		break
	}

	for _, swapInfo := range swapInfoList {
		status, err := TransferVerifyAtom(client, swapInfo)
		if err != nil {
			logrus.Errorf("atom TransferVerify failed: %s", err)
			return err
		}
		swapInfo.State = status
		err = dao_station.UpOrInFeeStationSwapInfo(db, swapInfo)
		if err != nil {
			logrus.Errorf("dao_station.UpOrInSwapInfo err: %s", err)
			return err
		}
	}
	return nil
}

func TransferVerifyAtom(client *cosmosRpc.Client, swapInfo *dao_station.FeeStationSwapInfo) (uint8, error) {
	stafiAddressBytes, err := hexutil.Decode(swapInfo.StafiAddress)
	if err != nil {
		return 0, err
	}
	blockHashBytes, err := hexutil.Decode(swapInfo.Blockhash)
	if err != nil {
		return 0, err
	}
	txHashBytes, err := hexutil.Decode(swapInfo.Txhash)
	if err != nil {
		return 0, err
	}
	pubkeyBytes, err := hexutil.Decode(swapInfo.Pubkey)
	if err != nil {
		return 0, err
	}
	poolAddr, err := types.AccAddressFromBech32(swapInfo.PoolAddress)
	if err != nil {
		return 0, err
	}

	inAmountDeci, err := decimal.NewFromString(swapInfo.InAmount)
	if err != nil {
		return 0, err
	}

	hashStr := hex.EncodeToString(txHashBytes)
	//check tx hash
	txRes, err := GetTx(client, hashStr)
	if err != nil {
		return utils.SwapStateTxHashFailed, nil
	}

	if txRes.Empty() {
		return utils.SwapStateTxHashFailed, nil
	}

	if txRes.Code != 0 {
		return utils.SwapStateTxHashFailed, nil
	}

	//check block hash
	blockRes, err := GetBlock(client, txRes.Height)
	if err != nil {
		return 0, err
	}
	if !bytes.Equal(blockRes.BlockID.Hash, blockHashBytes) {
		return utils.SwapStateBlockHashFailed, nil
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
		return utils.SwapStateAmountFailed, nil
	}
	if !poolIsMatch {
		return utils.SwapStatePoolAddressFailed, nil
	}

	//check pubkey
	fromAddress, err := types.AccAddressFromBech32(fromAddressStr)
	if err != nil {
		return 0, err
	}
	accountRes, err := client.QueryAccount(fromAddress)
	if err != nil {
		return 0, err
	}

	if !bytes.Equal(accountRes.GetPubKey().Bytes(), pubkeyBytes) {
		return utils.SwapStatePubkeyFailed, nil
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
		return 0, err
	}

	if memoInTx != bonderAddr {
		return utils.SwapStateMemoFailed, nil // memo unmatch
	}

	return utils.SwapStateVerifyTxOk, nil
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
