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
	SwapLimit    string     `json:"swapLimit"` //decimals 12
}

// @Summary get pool info
// @Description get pool info
// @Tags v1
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspPoolInfo}
// @Router /v1/station/poolInfo [get]
func (h *Handler) HandleGetPoolInfo(c *gin.Context) {
	list, err := dao_station.GetPoolAddressList(h.db)
	if err != nil {
		utils.Err(c, err.Error())
		return
	}
	swapRateStr := h.cache[utils.SwapRateKey]
	swapLimitStr := h.cache[utils.SwapLimitKey]
	swapRateDeci, err := decimal.NewFromString(swapRateStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,str:%s err %s", swapRateStr, err)
		swapRateDeci = defaultSwapRateDeci
	}
	swapLimitDeci, err := decimal.NewFromString(swapLimitStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,str:%s err %s", swapLimitStr, err)
		swapLimitDeci = defaultSwapLimitDeci
	}

	rsp := RspPoolInfo{
		PoolInfoList: make([]PoolInfo, 0),
		SwapLimit:    swapLimitDeci.StringFixed(0),
	}

	//get fis price
	fisPrice, err := dao_station.GetTokenPriceBySymbol(h.db, utils.SymbolFis)
	if err != nil {
		utils.Err(c, err.Error())
		return
	}
	fisPriceDeci, err := decimal.NewFromString(fisPrice.Price)
	if err != nil {
		utils.Err(c, err.Error())
		return
	}

	for _, l := range list {
		//get symbol price
		symbolPrice, err := dao_station.GetTokenPriceBySymbol(h.db, l.Symbol)
		if err != nil {
			utils.Err(c, err.Error())
			return
		}
		symbolPriceDeci, err := decimal.NewFromString(symbolPrice.Symbol)
		if err != nil {
			utils.Err(c, err.Error())
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
