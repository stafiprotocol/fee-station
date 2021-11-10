package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"math/big"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

//for native_chain_tx which's state == 0 and tx_status == 0:
//if native_chain_tx exist in swap_info:
//	if swap_info.state == 0:
//		skip
//	if swap_info.state == 1 or 2:
//		native_chain_tx.state = 1
//	if swap_info.state == 3 4 5 6 7 8:
//		if symbol == atom:
//			recover
//		else:
//			native_chain_tx.state = 1
//else:
//	if over recoverTime:
//		if found bondAddress:
//			recover
//		else:
//			if symbol == atom:
//				recover
//			else:
//				skip
//	else:
//		skip
func Recover(db *db.WrapDb, recoverTime, startTimestamp int64, swapMaxLimit string) error {
	nativeNotDealTxs, err := dao_station.GetFeeStationNativeTxByState(db, 0, 0, startTimestamp)
	if err != nil {
		return err
	}
	swapMaxLimitDeci, err := decimal.NewFromString(swapMaxLimit)
	if err != nil {
		return err
	}

	for _, tx := range nativeNotDealTxs {
		swapTx, err := dao_station.GetFeeStationSwapInfoByTx(db, tx.Txhash)
		if err == nil {
			switch swapTx.State {
			case 0:
				continue
			case 1, 2:
				tx.State = 1
				err = dao_station.UpOrInFeeStationNativeChainTx(db, tx)
				if err != nil {
					return err
				}
			case 3, 4, 5, 6, 7, 8:
				if strings.EqualFold(tx.Symbol, utils.SymbolAtom) {
					//recover
					if len(tx.Receiver) == 0 {
						logrus.Warnf("recover nativeTx, can't find stafiAddress in memo and will skip, nativeTx: %+v", *tx)
						tx.State = 1
						err = dao_station.UpOrInFeeStationNativeChainTx(db, tx)
						if err != nil {
							return err
						}
						continue
					}
					swapTx.Blockhash = tx.Blockhash
					swapTx.InAmount = tx.InAmount
					swapTx.StafiAddress = tx.Receiver
					swapTx.Pubkey = tx.SenderPubkey
					swapTx.PoolAddress = tx.PoolAddress
					swapTx.State = 0
					//in amount
					inAmountDeci, err := decimal.NewFromString(tx.InAmount)
					if err != nil {
						return err
					}
					symbolDecimals := utils.DecimalsMap[utils.SymbolAtom]
					realSwapRateDeci, err := decimal.NewFromString(swapTx.SwapRate)
					if err != nil {
						return err
					}
					//out amount
					outAmount := realSwapRateDeci.Mul(inAmountDeci).Div(decimal.NewFromBigInt(big.NewInt(1), symbolDecimals-6))
					if outAmount.Cmp(swapMaxLimitDeci) > 0 {
						outAmount = swapMaxLimitDeci
					}
					swapTx.OutAmount = outAmount.StringFixed(0)

					err = dao_station.UpOrInFeeStationSwapInfo(db, swapTx)
					if err != nil {
						return err
					}
				} else {
					tx.State = 1
					err = dao_station.UpOrInFeeStationNativeChainTx(db, tx)
					if err != nil {
						return err
					}
				}
			}

		} else {
			if err != gorm.ErrRecordNotFound {
				return err
			}
			//2 recover
			if tx.TxTimestamp+recoverTime < time.Now().Unix() {
				bundleAddressList, err := dao_station.GetFeeStationBundleAddressListByPubkeySymbol(db, tx.SenderPubkey, tx.Symbol)
				if err != nil {
					return err
				}

				//find bundle address
				var useBundleAddress *dao_station.FeeStationBundleAddress
				for _, bundleAddress := range bundleAddressList {
					if tx.TxTimestamp > int64(bundleAddress.CreatedAt) {
						useBundleAddress = bundleAddress
						break
					}
				}
				if useBundleAddress == nil {
					//if not found try memo
					if strings.EqualFold(tx.Symbol, utils.SymbolAtom) {
						if len(tx.Receiver) == 0 {
							tx.State = 1
							err = dao_station.UpOrInFeeStationNativeChainTx(db, tx)
							if err != nil {
								return err
							}
							logrus.Warnf("recover nativeTx, can't find stafiAddress in bundleAddress and memo and will skip, nativeTx: %+v", *tx)
							continue
						}

						okInfos, _ := dao_station.GetFeeStationSwapInfoListBySymbolState(db, utils.SymbolAtom, utils.SwapStatePayOk)
						if len(okInfos) == 0 {
							tx.State = 1
							err = dao_station.UpOrInFeeStationNativeChainTx(db, tx)
							if err != nil {
								return err
							}
							logrus.Warnf("recover nativeTx, can't find swaprate for atom and will skip, nativeTx: %+v", *tx)
							continue
						}

						swapTx.Blockhash = tx.Blockhash
						swapTx.Txhash = tx.Txhash
						swapTx.InAmount = tx.InAmount
						swapTx.StafiAddress = tx.Receiver
						swapTx.Pubkey = tx.SenderPubkey
						swapTx.PoolAddress = tx.PoolAddress
						swapTx.Symbol = tx.Symbol
						swapTx.State = 0
						//in amount
						inAmountDeci, err := decimal.NewFromString(tx.InAmount)
						if err != nil {
							return err
						}
						symbolDecimals := utils.DecimalsMap[utils.SymbolAtom]
						realSwapRateDeci, err := decimal.NewFromString(okInfos[0].SwapRate)
						if err != nil {
							return err
						}
						//out amount
						outAmount := realSwapRateDeci.Mul(inAmountDeci).Div(decimal.NewFromBigInt(big.NewInt(1), symbolDecimals-6))
						if outAmount.Cmp(swapMaxLimitDeci) > 0 {
							outAmount = swapMaxLimitDeci
						}
						swapTx.OutAmount = outAmount.StringFixed(0)
						swapTx.SwapRate = okInfos[0].SwapRate

						err = dao_station.UpOrInFeeStationSwapInfo(db, swapTx)
						if err != nil {
							return err
						}
					} else {
						tx.State = 1
						err = dao_station.UpOrInFeeStationNativeChainTx(db, tx)
						if err != nil {
							return err
						}
						logrus.Warnf("recover nativeTx, can't find stafiAddress in bundleAddress and will skip, nativeTx: %+v", *tx)
					}
					continue
				}

				//cal out amount
				realSwapRateDeci, err := decimal.NewFromString(useBundleAddress.SwapRate)
				if err != nil {
					return err
				}
				inAmountDeci, err := decimal.NewFromString(tx.InAmount)
				if err != nil {
					return err
				}

				symbolDecimals := utils.DecimalsMap[tx.Symbol]
				outAmount := realSwapRateDeci.Mul(inAmountDeci).Div(decimal.New(1, symbolDecimals-6))
				if outAmount.Cmp(swapMaxLimitDeci) > 0 {
					outAmount = swapMaxLimitDeci
				}

				newSwapInfo := dao_station.FeeStationSwapInfo{
					StafiAddress:    useBundleAddress.StafiAddress,
					State:           0,
					Symbol:          tx.Symbol,
					Blockhash:       tx.Blockhash,
					Txhash:          tx.Txhash,
					PoolAddress:     tx.PoolAddress,
					Signature:       "",
					Pubkey:          tx.SenderPubkey,
					InAmount:        tx.InAmount,
					MinOutAmount:    "",
					OutAmount:       outAmount.StringFixed(0),
					SwapRate:        useBundleAddress.SwapRate,
					InTokenPrice:    "",
					OutTokenPrice:   "",
					PayInfo:         "",
					BundleAddressId: useBundleAddress.ID,
				}

				err = dao_station.UpOrInFeeStationSwapInfo(db, &newSwapInfo)
				if err != nil {
					return err
				}
			}
		}

	}

	return nil

}
