package station_handlers

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/utils"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var defaultSwapLimitDeci = decimal.NewFromBigInt(big.NewInt(10), 12) //default 10e12
var defaultSwapRateDeci = decimal.NewFromBigInt(big.NewInt(1), 6)    //default 1e6
type ReqSwapInfo struct {
	StafiAddress string `json:"stafiAddress"` //hex
	Symbol       string `json:"symbol"`
	Blockhash    string `json:"blockHash"` //hex
	Txhash       string `json:"txHash"`    //hex
	PoolAddress  string `json:"poolAddress"`
	Signature    string `json:"signature"`    //hex
	Pubkey       string `json:"pubkey"`       //hex
	InAmount     string `json:"inAmount"`     //decimal
	MinOutAmount string `json:"minOutAmount"` //decimal
}

// @Summary post swap info
// @Description post swap info
// @Tags v1
// @Accept json
// @Produce json
// @Param param body ReqSwapInfo true "user swap info"
// @Success 200 {object} utils.Rsp{}
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
	//check 0x prefix param
	var stafiAddressBytes []byte
	if stafiAddressBytes, err = hexutil.Decode(req.StafiAddress); err != nil {
		utils.Err(c, "stafiAddress format err")
		return
	}
	if _, err := hexutil.Decode(req.Blockhash); err != nil {
		utils.Err(c, "blockHash format err")
		return
	}
	if _, err := hexutil.Decode(req.Txhash); err != nil {
		utils.Err(c, "txHash format err")
		return
	}

	var sigBytes []byte
	var pubkeyBytes []byte
	if sigBytes, err = hexutil.Decode(req.Signature); req.Symbol != utils.SymbolAtom && err != nil {
		utils.Err(c, "signature format err")
		return
	}
	if pubkeyBytes, err = hexutil.Decode(req.Pubkey); err != nil {
		utils.Err(c, "pubkey format err")
		return
	}

	//check pool address
	poolAddr, err := dao_station.GetPoolAddressBySymbol(h.db, req.Symbol)
	if err != nil {
		utils.Err(c, "get pool address failed")
		logrus.Errorf("dao_station.GetPoolAddressBySymbol err %v", err)
		return
	}
	if !strings.EqualFold(poolAddr.PoolAddress, req.PoolAddress) {
		utils.Err(c, "pool address not right")
		logrus.Errorf("pool address not right:req %s,db:%s", req.PoolAddress, poolAddr.PoolAddress)
		return
	}

	//check block hash and tx hash duplicate
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

	//check signature
	switch req.Symbol {
	case utils.SymbolDot, utils.SymbolKsm:
		ok := utils.VerifySigsSecp256(sigBytes, pubkeyBytes, stafiAddressBytes)
		if !ok {
			utils.Err(c, "signature not right")
			logrus.Errorf("utils.VerifySigsSecp256 failed, stafi address: %s", req.StafiAddress)
			return
		}
	case utils.SymbolEth:
		ok := utils.VerifySigsEth(sigBytes, stafiAddressBytes, common.BytesToAddress(pubkeyBytes))
		if !ok {
			utils.Err(c, "signature not right")
			logrus.Errorf("utils.VerifySigsEth failed, stafi address: %s", req.StafiAddress)
			return
		}
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
	//get symbol price
	symbolPrice, err := dao_station.GetTokenPriceBySymbol(h.db, req.Symbol)
	if err != nil {
		utils.Err(c, err.Error())
		return
	}
	symbolPriceDeci, err := decimal.NewFromString(symbolPrice.Price)
	if err != nil {
		utils.Err(c, err.Error())
		return
	}
	//swap rate
	swapRateStr := h.cache[utils.SwapRateKey]
	swapLimitStr := h.cache[utils.SwapLimitKey]
	swapRateDeci, err := decimal.NewFromString(swapRateStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapRateStr: %s err %s", swapRateStr, err)
		swapRateDeci = defaultSwapRateDeci
	}
	swapLimitDeci, err := decimal.NewFromString(swapLimitStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapLimitStr: %s err %s", swapLimitStr, err)
		swapLimitDeci = defaultSwapLimitDeci
	}

	//cal real swap rate
	realSwapRateDeci := swapRateDeci.Mul(symbolPriceDeci).Div(fisPriceDeci)
	//in amount
	inAmountDeci, err := decimal.NewFromString(req.InAmount)
	if err != nil {
		utils.Err(c, err.Error())
		return
	}
	//out amount
	outAmount := realSwapRateDeci.Mul(inAmountDeci).Div(decimal.NewFromBigInt(big.NewInt(1), 12))
	if outAmount.Cmp(swapLimitDeci) > 0 {
		outAmount = swapLimitDeci
	}
	//check min out amount
	minOutAmountDeci, err := decimal.NewFromString(req.MinOutAmount)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,minOutAmount: %s err %s", req.MinOutAmount, err)
		utils.Err(c, err.Error())
		return
	}
	if outAmount.Cmp(minOutAmountDeci) < 0 {
		utils.Err(c, "real out amount < min out amount")
		return
	}

	swapInfo.StafiAddress = strings.ToLower(req.StafiAddress)
	swapInfo.Symbol = req.Symbol
	swapInfo.Blockhash = strings.ToLower(req.Blockhash)
	swapInfo.Txhash = strings.ToLower(req.Txhash)
	swapInfo.PoolAddress = req.PoolAddress
	swapInfo.Signature = strings.ToLower(req.Signature)
	swapInfo.Pubkey = strings.ToLower(req.Pubkey)
	swapInfo.InAmount = req.InAmount
	swapInfo.MinOutAmount = req.MinOutAmount
	swapInfo.SwapRate = realSwapRateDeci.StringFixed(0)
	swapInfo.OutAmount = outAmount.StringFixed(0)
	swapInfo.State = utils.SwapStateVerifySigs

	//update db
	err = dao_station.UpOrInSwapInfo(h.db, swapInfo)
	if err != nil {
		utils.Err(c, err.Error())
		logrus.Errorf("UpOrInSwapInfo err %v", err)
		return
	}

	utils.Ok(c, "success", struct{}{})
}

type RspSwapInfo struct {
	SwapStatus uint8 `json:"swapStatus"`
}

// @Summary get swap info
// @Description get swap info
// @Tags v1
// @Param symbol query string true "token symbol"
// @Param blockHash query string true "block hash hex string"
// @Param txHash query string true "tx hash hex string"
// @Produce json
// @Success 200 {object} utils.Rsp{data=RspSwapInfo}
// @Router /v1/station/swapInfo [get]
func (h *Handler) HandleGetSwapInfo(c *gin.Context) {
	symbol := c.Query("symbol")
	blockHash := c.Query("blockHash")
	txHash := c.Query("txHash")
	//check param
	if !utils.SymbolValid(symbol) {
		utils.Err(c, "symbol unsupport")
		return
	}
	if _, err := hexutil.Decode(blockHash); err != nil {
		utils.Err(c, "blockHash format err")
		return
	}
	if _, err := hexutil.Decode(txHash); err != nil {
		utils.Err(c, "txHash format err")
		return
	}

	swapInfo, err := dao_station.GetSwapInfoBySymbolBlkTx(h.db, symbol, strings.ToLower(blockHash), strings.ToLower(txHash))
	if err != nil {
		utils.Err(c, err.Error())
		return
	}
	rsp := RspSwapInfo{
		SwapStatus: swapInfo.State,
	}
	utils.Ok(c, "success", rsp)
}
