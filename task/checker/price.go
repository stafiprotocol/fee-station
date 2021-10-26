package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fmt"
	"math/big"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

func UpdatePrice(db *db.WrapDb, coinMarketApi, coinGeckoApi string) error {
	coinMarketUrl := fmt.Sprintf("%s?symbol=%s,%s,%s,%s,%s", coinMarketApi,
		utils.SymbolAtom, utils.SymbolDot, utils.SymbolEth, utils.SymbolKsm, utils.SymbolFis)
	coinGeckoUrl := fmt.Sprintf("%s?ids=ethereum,polkadot,cosmos,stafi,kusama&vs_currencies=usd", coinGeckoApi)

	retry := 0
	var priceMap map[string]float64
	var err error
	for {
		if retry > BlockRetryLimit {
			return fmt.Errorf("cosmosRpc.NewClient reach retry limit")
		}

		priceMap, err = utils.GetPriceFromCoinMarket(coinMarketUrl)
		if err != nil {
			logrus.Warnf("GetPriceFromCoinMarket err: %s, will try coinGecko", err)
			priceMap, err = utils.GetPriceFromCoinGecko(coinGeckoUrl)
			if err != nil {
				logrus.Warnf("GetPriceFromCoinGecko err: %s, will retry", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue
			}
		}
		break
	}

	for key, value := range priceMap {
		token, _ := dao_station.GetFeeStationTokenPriceBySymbol(db, key)
		token.Symbol = key
		token.Price = decimal.NewFromFloat(value).Mul(decimal.NewFromBigInt(big.NewInt(1), 18)).StringFixed(0)
		err := dao_station.UpOrInFeeStationTokenPrice(db, token)
		if err != nil {
			return err
		}
	}
	return nil
}
