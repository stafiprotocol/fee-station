package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fee-station/shared/substrate"

	"github.com/sirupsen/logrus"
)

func CheckKsmTx(db *db.WrapDb, ksmEndpoint, typesPath string) error {
	swapInfoList, err := dao_station.GetSwapInfoListBySymbolState(db, utils.SymbolKsm, utils.SwapStateVerifySigs)
	if err != nil {
		return err
	}

	sc, err := substrate.NewSarpcClient(substrate.ChainTypePolkadot, ksmEndpoint, typesPath)
	if err != nil {
		return err
	}
	gc, err := substrate.NewGsrpcClient(ksmEndpoint, substrate.AddressTypeAccountId, nil)
	if err != nil {
		return err
	}

	for _, swapInfo := range swapInfoList {
		status, err := TransferVerifySubstrate(gc, sc, swapInfo)
		if err != nil {
			logrus.Errorf("dot TransferVerify failed: %s", err)
			continue
		}
		swapInfo.State = status
		err = dao_station.UpOrInSwapInfo(db, swapInfo)
		if err != nil {
			logrus.Warnf("dao_station.UpOrInSwapInfo err: %s", err)
			continue
		}
	}
	return nil
}
