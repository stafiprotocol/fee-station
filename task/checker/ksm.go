package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fee-station/shared/substrate"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func CheckKsmTx(db *db.WrapDb, ksmEndpoint, typesPath string) error {
	swapInfoList, err := dao_station.GetFeeStationSwapInfoListBySymbolState(db, utils.SymbolKsm, utils.SwapStateVerifySigs)
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
		sc, err = substrate.NewSarpcClient(substrate.ChainTypePolkadot, ksmEndpoint, typesPath)
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
		gc, err = substrate.NewGsrpcClient(ksmEndpoint, substrate.AddressTypeAccountId, nil)
		if err != nil {
			logrus.Warnf("substrate.NewSarpcClient err: %s", err)
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
		err = dao_station.UpOrInFeeStationSwapInfo(db, swapInfo)
		if err != nil {
			logrus.Errorf("dao_station.UpOrInSwapInfo err: %s", err)
			return err
		}
	}
	return nil
}
