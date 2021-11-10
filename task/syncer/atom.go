package task

import (
	"bytes"
	"errors"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	cosmosRpc "fee-station/shared/cosmos/rpc"
	"fmt"
	"strings"
	"time"

	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/cosmos/cosmos-sdk/types"
	xBankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	cTypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/sirupsen/logrus"
)

var (
	pageLimit = 10
	dayLayout = "2006-01-02T15:04:05Z"
)

func SyncAtomTx(db *db.WrapDb, denom, atomEndpoint string) error {
	poolAddressRes, err := dao_station.GetFeeStationPoolAddressBySymbol(db, utils.SymbolAtom)
	if err != nil {
		return err
	}
	poolAddress := poolAddressRes.PoolAddress
	poolAddr, err := types.AccAddressFromBech32(poolAddress)
	if err != nil {
		return err
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
	filter := []string{fmt.Sprintf("transfer.recipient='%s'", poolAddress), "message.module='bank'"}

	for {
		totalCount, err := dao_station.GetFeeStationNativeChainTxTotalCount(db, utils.SymbolAtom)
		if err != nil {
			return err
		}
		txResPre, err := client.GetEvents(filter, int(1), pageLimit, "asc")
		if err != nil {
			return err
		}
		usePage := totalCount/int64(pageLimit) + 1

		//sip if localdb have
		if uint64(usePage) > txResPre.PageTotal {
			return nil
		}

		txRes, err := client.GetEvents(filter, int(usePage), pageLimit, "asc")
		if err != nil {
			return err
		}

		for _, tx := range txRes.Txs {
			useTxHash := strings.ToLower("0x" + tx.TxHash)

			_, err := dao_station.GetFeeStationNativeChainTxBySymbolTxhash(db, utils.SymbolAtom, useTxHash)
			//skip if exist
			if err == nil {
				continue
			}

			resBlock, err := GetBlock(client, tx.Height)
			if err != nil {
				return err
			}

			senderAddr := ""
			inAmount := ""
			msgs := tx.GetTx().GetMsgs()

		out:
			for i, _ := range msgs {
				switch msgs[i].Type() {
				case xBankTypes.TypeMsgSend:
					if sendMsg, ok := msgs[i].(*xBankTypes.MsgSend); ok {
						toAddr, err := types.AccAddressFromBech32(sendMsg.ToAddress)
						if err == nil {
							//amount and pool address must in one message
							if bytes.Equal(toAddr.Bytes(), poolAddr.Bytes()) {
								inAmount = sendMsg.Amount.AmountOf(client.GetDenom()).String()
								senderAddr = sendMsg.FromAddress
								break out
							}
						}
					}
				case xBankTypes.TypeMsgMultiSend: //not support, will skip
					if multiSendMsg, ok := msgs[i].(*xBankTypes.MsgMultiSend); ok {
						for _, output := range multiSendMsg.Outputs {
							toAddr, err := types.AccAddressFromBech32(output.Address)
							if err == nil {
								//amount and pool address must in one message
								if bytes.Equal(toAddr.Bytes(), poolAddr.Bytes()) {
									inAmount = output.Coins.AmountOf(client.GetDenom()).String()
									senderAddr = multiSendMsg.Inputs[0].Address
									break out
								}
							}
						}

					}
				}
			}

			if len(inAmount) == 0 {
				return fmt.Errorf("get amount failed,msgs: %v", msgs)
			}

			//get pubkey
			fromAddress, err := types.AccAddressFromBech32(senderAddr)
			if err != nil {
				return err
			}
			accountRes, err := client.QueryAccount(fromAddress)
			if err != nil {
				return err
			}
			txTimestamp, err := time.Parse(dayLayout, tx.Timestamp)
			if err != nil {
				return err
			}
			//got receiver
			var receiver string
			var memoInTx string
			txx, err := client.GetTxConfig().TxDecoder()(tx.Tx.GetValue())
			if err == nil {
				memoTx, ok := txx.(types.TxWithMemo)
				if ok {
					memoInTx = memoTx.GetMemo()
				}
			}
			if len(memoInTx) != 0 {
				receiverBts, err := ss58.DecodeToPub(memoInTx)
				if err == nil {
					receiver = strings.ToLower(hexutil.Encode(receiverBts))
				}
			}

			usePubkey := strings.ToLower(hexutil.Encode(accountRes.GetPubKey().Bytes()))
			txStatus := int64(0)
			if tx.Code != 0 {
				txStatus = int64(tx.Code)
			}

			nativeTx := dao_station.FeeStationNativeChainTx{
				State:        0,
				TxStatus:     txStatus,
				Symbol:       utils.SymbolAtom,
				Blockhash:    strings.ToLower(hexutil.Encode(resBlock.BlockID.Hash.Bytes())),
				Txhash:       useTxHash,
				PoolAddress:  poolAddress,
				SenderPubkey: usePubkey,
				InAmount:     inAmount,
				TxTimestamp:  txTimestamp.Unix(),
				Receiver:     receiver,
			}

			err = dao_station.UpOrInFeeStationNativeChainTx(db, &nativeTx)
			if err != nil {
				return err
			}
		}

		//just break when get all
		if txRes.PageTotal == txRes.PageNumber {
			break
		}
	}

	return nil
}

func GetBlock(client *cosmosRpc.Client, height int64) (*cTypes.ResultBlock, error) {
	var blockRes *cTypes.ResultBlock
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
