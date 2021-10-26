package station_handlers

import (
	"fee-station/dao/station"
	"fee-station/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type PoolInfo struct {
	Symbol      string `json:"symbol"`
	PoolAddress string `json:"poolAddress"` //base58 or hex
	SwapRate    string `json:"swapRate"`    //decimals 6
}

type RspPoolInfo struct {
	PoolInfoList []PoolInfo `json:"poolInfoList"`
	SwapMaxLimit string     `json:"swapMaxLimit"` //decimals 12
	SwapMinLimit string     `json:"swapMinLimit"` //decimals 12
}

// @Summary get pool info
// @Description get pool info
// @Tags v1
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspPoolInfo}
// @Router /v1/station/poolInfo [get]
func (h *Handler) HandleGetPoolInfo(c *gin.Context) {
	list, err := dao_station.GetFeeStationPoolAddressList(h.db)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		return
	}
	swapRateStr := h.cache[utils.SwapRateKey]
	swapMaxLimitStr := h.cache[utils.SwapMaxLimitKey]
	swapMinLimitStr := h.cache[utils.SwapMinLimitKey]
	swapRateDeci, err := decimal.NewFromString(swapRateStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,str:%s err %s", swapRateStr, err)
		swapRateDeci = defaultSwapRateDeci
	}
	swapMaxLimitDeci, err := decimal.NewFromString(swapMaxLimitStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapMaxLimitStr:%s err %s", swapMaxLimitStr, err)
		swapMaxLimitDeci = defaultSwapMaxLimitDeci
	}
	swapMinLimitDeci, err := decimal.NewFromString(swapMinLimitStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapMinLimitStr:%s err %s", swapMinLimitStr, err)
		swapMinLimitDeci = defaultSwapMinLimitDeci
	}

	rsp := RspPoolInfo{
		PoolInfoList: make([]PoolInfo, 0),
		SwapMaxLimit: swapMaxLimitDeci.StringFixed(0),
		SwapMinLimit: swapMinLimitDeci.StringFixed(0),
	}

	//get fis price
	fisPrice, err := dao_station.GetFeeStationTokenPriceBySymbol(h.db, utils.SymbolFis)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}
	fisPriceDeci, err := decimal.NewFromString(fisPrice.Price)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}

	for _, l := range list {
		//get symbol price
		symbolPrice, err := dao_station.GetFeeStationTokenPriceBySymbol(h.db, l.Symbol)
		if err != nil {
			utils.Err(c, codeTokenPriceErr, err.Error())
			return
		}
		symbolPriceDeci, err := decimal.NewFromString(symbolPrice.Price)
		if err != nil {
			utils.Err(c, codeTokenPriceErr, err.Error())
			return
		}
		//cal real swap rate
		realSwapRateDeci := swapRateDeci.Mul(symbolPriceDeci).Div(fisPriceDeci)

		rsp.PoolInfoList = append(rsp.PoolInfoList, PoolInfo{
			Symbol:      l.Symbol,
			PoolAddress: l.PoolAddress,
			SwapRate:    realSwapRateDeci.StringFixed(0),
		})
	}

	utils.Ok(c, "success", rsp)
}
