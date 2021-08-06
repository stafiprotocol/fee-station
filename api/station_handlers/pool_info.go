package station_handlers

import (
	"fee-station/dao/station"
	"fee-station/pkg/utils"

	"github.com/gin-gonic/gin"
)

type PoolInfo struct {
	Symbol      string `json:"symbol"`
	PoolAddress string `json:"poolAddress"`
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
	rsp := RspPoolInfo{
		PoolInfoList: make([]PoolInfo, 0),
	}
	for _, l := range list {
		rsp.PoolInfoList = append(rsp.PoolInfoList, PoolInfo{
			Symbol:      l.Symbol,
			PoolAddress: l.PoolAddress,
		})
	}

	utils.Ok(c, "success", rsp)
}
