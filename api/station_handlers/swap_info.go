package station_handlers

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReqSwapInfo struct {
	StafiAddress string `json:"stafiAddress"`
	Symbol       string `json:"symbol"`
	Blockhash    string `json:"blockHash"`
	Txhash       string `json:"txHash"`
	PoolAddress  string `json:"poolAddress"`
	Signature    string `json:"signature"`
	Pubkey       string `json:"pubkey"`
	InAmount     string `json:"inAmount"`
	MinOutAmount string `json:"minOutAmount"`
}

// @Summary post swap info
// @Description post swap info
// @Tags v1
// @Accept json
// @Produce json
// @Param param body ReqSwapInfo true "user swap info"
// @Success 200 {object} utils.Rsp{data=RspPoolInfo}
// @Router /v1/station/swapInfo [post]
func (h *Handler) HandlePostSwapInfo(c *gin.Context) {
	req := ReqSwapInfo{}
	err := c.Bind(&req)
	if err != nil {
		utils.Err(c, err.Error())
		logrus.Errorf("bind err %v", err)
		return
	}
	//check symbol
	if !utils.SymbolValid(req.Symbol) {
		utils.Err(c, "symbol unsupport")
		return
	}

	//check duplicate
	swapInfo, err := dao_station.GetSwapInfoBySymbolBlkTx(
		h.db, req.Symbol, strings.ToLower(req.Blockhash), strings.ToLower(req.Txhash))
	if err != nil && err != gorm.ErrRecordNotFound {
		utils.Err(c, err.Error())
		logrus.Errorf("GetSwapInfoBySymbolBlkTx err %v", err)
		return
	}
	if err == nil {
		utils.Err(c, "duplicate swap info")
		logrus.Errorf("duplicate swap info, txhash:", req.Txhash)
		return
	}

	swapInfo.StafiAddress = req.StafiAddress
	swapInfo.Symbol = req.Symbol
	swapInfo.Blockhash = strings.ToLower(req.Blockhash)
	swapInfo.Txhash = strings.ToLower(req.Txhash)
	swapInfo.PoolAddress = req.PoolAddress
	swapInfo.Signature = req.Signature
	swapInfo.Pubkey = req.Pubkey
	swapInfo.InAmount = req.InAmount
	swapInfo.MinOutAmount = req.MinOutAmount

	//update db
	err = dao_station.UpOrInSwapInfo(h.db, swapInfo)
	if err != nil {
		utils.Err(c, err.Error())
		logrus.Errorf("UpOrInSwapInfo err %v", err)
		return
	}

	utils.Ok(c, "success", struct{}{})
}
