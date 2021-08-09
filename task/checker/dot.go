package task

import (
	"errors"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fee-station/shared/substrate"
	"fmt"
	"strings"
	"time"

	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/itering/scale.go/utiles"
	"github.com/sirupsen/logrus"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

func CheckDotTx(db *db.WrapDb, dotEndpoint, typesPath string) error {
	swapInfoList, err := dao_station.GetSwapInfoListBySymbolState(db, utils.SymbolDot, utils.SwapStateVerifySigs)
	if err != nil {
		return err
	}
	if len(swapInfoList) == 0 {
		return nil
	}

	retry := 0
	var sc *substrate.SarpcClient
	for {
		if retry > BlockRetryLimit {
			return fmt.Errorf("substrate.NewSarpcClient reach retry limit")
		}
		sc, err = substrate.NewSarpcClient(substrate.ChainTypePolkadot, dotEndpoint, typesPath)
		if err != nil {
			logrus.Warnf("substrate.NewSarpcClient err: %s", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		break
	}

	retry = 0
	var gc *substrate.GsrpcClient
	for {
		if retry > BlockRetryLimit {
			return fmt.Errorf("substrate.NewGsrpcClient reach retry limit")
		}
		gc, err = substrate.NewGsrpcClient(dotEndpoint, substrate.AddressTypeAccountId, nil)
		if err != nil {
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		break
	}

	for _, swapInfo := range swapInfoList {
		status, err := TransferVerifySubstrate(gc, sc, swapInfo)
		if err != nil {
			logrus.Errorf("dot TransferVerify failed: %s", err)
			return err
		}
		swapInfo.State = status
		err = dao_station.UpOrInSwapInfo(db, swapInfo)
		if err != nil {
			logrus.Errorf("dao_station.UpOrInSwapInfo err: %s", err)
			return err
		}
	}

	return nil
}

func TransferVerifySubstrate(gc *substrate.GsrpcClient, sc *substrate.SarpcClient, swapInfo *dao_station.SwapInfo) (uint8, error) {
	bh := swapInfo.Blockhash
	hash, err := types.NewHashFromHexString(swapInfo.Blockhash)
	if err != nil {
		return 0, err
	}
	poolAddressByte, err := ss58.Decode(swapInfo.PoolAddress)
	if err != nil {
		return 0, err
	}
	poolAddressPubkeyHexStr := hexutil.Encode(poolAddressByte[1:33])

	blkNum, err := gc.GetBlockNumber(hash)
	if err != nil {
		return 0, err
	}

	if blkNum == 0 {
		for i := 0; i < 10; i++ {
			time.Sleep(BlockRetryInterval)
			blkNum, err = gc.GetBlockNumber(hash)
			if err != nil {
				return 0, err
			}
			if blkNum != 0 {
				break
			}
		}
		if blkNum == 0 {
			return utils.SwapStateBlockHashFailed, nil
		}
	}

	final, err := gc.GetFinalizedBlockNumber()
	if err != nil {
		return 0, err
	}

	if blkNum > final {
		logrus.Info("TransferVerify: block hash not finalized, waiting", "blockHash", bh, "symbol", swapInfo.Symbol)
		time.Sleep(10 * BlockRetryInterval)
		final, err = gc.GetFinalizedBlockNumber()
		if err != nil {
			return 0, err
		}
		if blkNum > final {
			return 0, errors.New("block number not finalized")
		}
	}

	exts, err := sc.GetExtrinsics(bh)
	if err != nil {
		logrus.Warn("TransferVerify: get extrinsics error", "err", err, "blockHash", bh)
		return 0, err
	}

	th := swapInfo.Txhash
	for _, ext := range exts {
		logrus.Info("TransferVerify loop extrinsics", "ext", ext)
		txhash := utiles.AddHex(ext.ExtrinsicHash)
		if th != txhash {
			logrus.Info("txhash not equal", "expected", th, "got", txhash)
			continue
		}
		logrus.Info("txhash equal", "expected", th, "got", txhash)
		logrus.Info("TransferVerify", "CallModuleName", ext.CallModuleName, "CallName", ext.CallName, "ext.Params number", len(ext.Params))

		if ext.CallModuleName != substrate.BalancesModuleId || (ext.CallName != substrate.TransferKeepAlive && ext.CallName != substrate.Transfer) {
			return utils.SwapStateTxHashFailed, nil
		}

		addr, ok := ext.Address.(string)
		if !ok {
			logrus.Warn("TransferVerify: address not string", "address", ext.Address)
			return utils.SwapStatePubkeyFailed, nil
		}

		if !strings.EqualFold(swapInfo.Pubkey, utiles.AddHex(addr)) {
			logrus.Warn("TransferVerify: pubkey", "addr", addr, "pubkey", swapInfo.Pubkey)
			return utils.SwapStatePubkeyFailed, nil
		}

		for _, p := range ext.Params {
			logrus.Info("TransferVerify", "name", p.Name, "type", p.Type)
			if p.Name == substrate.ParamDest && p.Type == substrate.ParamDestType {
				logrus.Debug("cmp dest", "pool", swapInfo.PoolAddress, "dest", p.Value)

				dest, ok := p.Value.(string)
				if !ok {
					dest, ok := p.Value.(map[string]interface{})
					if !ok {
						return utils.SwapStatePoolAddressFailed, nil
					}

					destId, ok := dest["Id"]
					if !ok {
						return utils.SwapStatePoolAddressFailed, nil
					}

					d, ok := destId.(string)
					if !ok {
						return utils.SwapStatePoolAddressFailed, nil
					}

					if !strings.EqualFold(poolAddressPubkeyHexStr, utiles.AddHex(d)) {
						return utils.SwapStatePoolAddressFailed, nil
					}
				} else {
					if !strings.EqualFold(poolAddressPubkeyHexStr, utiles.AddHex(dest)) {
						return utils.SwapStatePoolAddressFailed, nil
					}
				}
			} else if p.Name == substrate.ParamValue && p.Type == substrate.ParamValueType {
				logrus.Info("cmp amount", "amount", swapInfo.InAmount, "paramAmount", p.Value)
				if fmt.Sprint(swapInfo.InAmount) != fmt.Sprint(p.Value) {
					return utils.SwapStateAmountFailed, nil
				}
			} else {
				logrus.Error("TransferVerify unexpected param", "name", p.Name, "value", p.Value, "type", p.Type)
				return utils.SwapStateTxHashFailed, nil
			}
		}

		return utils.SwapStateVerifyTxOk, nil
	}

	return utils.SwapStateTxHashFailed, nil
}

// 222
// 0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a488f0a
// 0x8eaf04151687736326c9fea17e25fc5287613693c912909cb226aa4794f26a48
