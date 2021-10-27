package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

//1 tag native_chain_tx stake=1 which exist in swap_info
//2 find native_chain_tx need recover and insert to swap_info
func Recover(db *db.WrapDb, recoverTime int64, swapMaxLimit string) error {
	nativeNotDealTxs, err := dao_station.GetFeeStationNativeTxByState(db, 0, 0)
	if err != nil {
		return err
	}
	swapMaxLimitDeci, err := decimal.NewFromString(swapMaxLimit)
	if err != nil {
		return err
	}

	for _, tx := range nativeNotDealTxs {
		_, err := dao_station.GetFeeStationSwapInfoByTx(db, tx.Txhash)
		//1 tag this tx is dealed if it exist in swap_info
		if err == nil {
			tx.State = 1
			err = dao_station.UpOrInFeeStationNativeChainTx(db, tx)
			if err != nil {
				return err
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
					logrus.Warnf("recover nativeTx,can't find stafiAddress and will skip, nativeTx: %+v", *tx)
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
