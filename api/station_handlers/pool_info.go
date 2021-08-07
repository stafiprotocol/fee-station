package station_handlers

import (
	"fee-station/dao/station"
	"fee-station/pkg/utils"
	"math/big"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type PoolInfo struct {
	Symbol      string `json:"symbol"`
	PoolAddress string `json:"poolAddress"` //base58 or hex
	SwapRate    string `json:"swapRate"`    //decimals 6
	SwapLimit   string `json:"swapLimit"`   //decimals 12
}

type RspPoolInfo struct {
	PoolInfoList []PoolInfo `json:"poolInfoList"`
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
		swapRateDeci = decimal.NewFromBigInt(big.NewInt(1), 6) //default 1e6
	}
	swapLimitDeci, err := decimal.NewFromString(swapLimitStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,str:%s err %s", swapLimitStr, err)
		swapLimitDeci = decimal.NewFromBigInt(big.NewInt(10), 12) //default 10e12
	}

	rsp := RspPoolInfo{
		PoolInfoList: make([]PoolInfo, 0),
	}
	for _, l := range list {
		rsp.PoolInfoList = append(rsp.PoolInfoList, PoolInfo{
			Symbol:      l.Symbol,
			PoolAddress: l.PoolAddress,
			SwapRate:    swapRateDeci.StringFixed(0),
			SwapLimit:   swapLimitDeci.StringFixed(0),
		})
	}

	utils.Ok(c, "success", rsp)
}
