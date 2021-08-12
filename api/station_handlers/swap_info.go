package station_handlers

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/utils"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var priceExpiredSeconds = 60 * 60 * 24 * 3 // 3 days

var defaultSwapMaxLimitDeci = decimal.NewFromBigInt(big.NewInt(100), 12) //default 100e12
var defaultSwapMinLimitDeci = decimal.NewFromBigInt(big.NewInt(1), 12)   //default 1e12
var defaultSwapRateDeci = decimal.NewFromBigInt(big.NewInt(1), 6)        //default 1e6
var decimalsMap = map[string]int32{
	utils.SymbolAtom: 6,
	utils.SymbolDot:  10,
	utils.SymbolKsm:  12,
	utils.SymbolEth:  18,
}

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
		utils.Err(c, codeParamParseErr, err.Error())
		logrus.Errorf("bind err %v", err)
		return
	}
	//check symbol
	if !utils.SymbolValid(req.Symbol) {
		utils.Err(c, codeSymbolErr, "symbol unsupport")
		return
	}
	//check 0x prefix param
	var stafiAddressBytes []byte
	if stafiAddressBytes, err = hexutil.Decode(req.StafiAddress); err != nil {
		utils.Err(c, codeStafiAddressErr, "stafiAddress format err")
		return
	}
	if len(stafiAddressBytes) != 32 {
		utils.Err(c, codeStafiAddressErr, "stafiAddress err")
		return
	}
	if _, err := hexutil.Decode(req.Blockhash); err != nil {
		utils.Err(c, codeBlockHashErr, "blockHash format err")
		return
	}
	if _, err := hexutil.Decode(req.Txhash); err != nil {
		utils.Err(c, codeTxHashErr, "txHash format err")
		return
	}

	var sigBytes []byte
	var pubkeyBytes []byte
	if sigBytes, err = hexutil.Decode(req.Signature); req.Symbol != utils.SymbolAtom && err != nil {
		utils.Err(c, codeSignatureErr, "signature format err")
		return
	}
	if pubkeyBytes, err = hexutil.Decode(req.Pubkey); err != nil {
		utils.Err(c, codePubkeyErr, "pubkey format err")
		return
	}

	//check pool address
	poolAddr, err := dao_station.GetPoolAddressBySymbol(h.db, req.Symbol)
	if err != nil {
		utils.Err(c, codeInternalErr, "get pool address failed")
		logrus.Errorf("dao_station.GetPoolAddressBySymbol err %v", err)
		return
	}
	if !strings.EqualFold(poolAddr.PoolAddress, req.PoolAddress) {
		utils.Err(c, codePoolAddressErr, "pool address not right")
		logrus.Errorf("pool address not right:req %s,db:%s", req.PoolAddress, poolAddr.PoolAddress)
		return
	}

	//check block hash and tx hash duplicate
	swapInfo, err := dao_station.GetSwapInfoBySymbolBlkTx(
		h.db, req.Symbol, strings.ToLower(req.Blockhash), strings.ToLower(req.Txhash))
	if err != nil && err != gorm.ErrRecordNotFound {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("GetSwapInfoBySymbolBlkTx err %v", err)
		return
	}
	if err == nil {
		utils.Err(c, codeTxDuplicateErr, "duplicate swap info")
		logrus.Errorf("duplicate swap info, txhash:", req.Txhash)
		return
	}

	//check signature
	switch req.Symbol {
	case utils.SymbolDot, utils.SymbolKsm:
		ok := utils.VerifiySigsSr25519(sigBytes, pubkeyBytes, stafiAddressBytes)
		if !ok {
			utils.Err(c, codeSignatureErr, "signature not right")
			logrus.Errorf("utils.VerifySigsSecp256 failed, stafi address: %s", req.StafiAddress)
			return
		}
	case utils.SymbolEth:
		ok := utils.VerifySigsEth(sigBytes, stafiAddressBytes, common.BytesToAddress(pubkeyBytes))
		if !ok {
			utils.Err(c, codeSignatureErr, "signature not right")
			logrus.Errorf("utils.VerifySigsEth failed, stafi address: %s", req.StafiAddress)
			return
		}
	}
	//get fis price
	fisPrice, err := dao_station.GetTokenPriceBySymbol(h.db, utils.SymbolFis)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}
	//check old price
	duration := int(time.Now().Unix()) - fisPrice.UpdatedAt
	if duration > priceExpiredSeconds {
		utils.Err(c, codeTokenPriceErr, "price too old")
		return
	}

	fisPriceDeci, err := decimal.NewFromString(fisPrice.Price)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}
	//get symbol price
	symbolPrice, err := dao_station.GetTokenPriceBySymbol(h.db, req.Symbol)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}
	symbolPriceDeci, err := decimal.NewFromString(symbolPrice.Price)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		return
	}
	//swap rate
	swapRateStr := h.cache[utils.SwapRateKey]
	swapMaxLimitStr := h.cache[utils.SwapMaxLimitKey]
	swapMinLimitStr := h.cache[utils.SwapMinLimitKey]
	swapRateDeci, err := decimal.NewFromString(swapRateStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapRateStr: %s err %s", swapRateStr, err)
		swapRateDeci = defaultSwapRateDeci
	}
	swapMaxLimitDeci, err := decimal.NewFromString(swapMaxLimitStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapMaxLimitStr: %s err %s", swapMaxLimitStr, err)
		swapMaxLimitDeci = defaultSwapMaxLimitDeci
	}
	swapMinLimitDeci, err := decimal.NewFromString(swapMinLimitStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapMinLimitStr: %s err %s", swapMinLimitStr, err)
		swapMinLimitDeci = defaultSwapMinLimitDeci
	}

	//cal real swap rate
	realSwapRateDeci := swapRateDeci.Mul(symbolPriceDeci).Div(fisPriceDeci)
	//in amount
	inAmountDeci, err := decimal.NewFromString(req.InAmount)
	if err != nil {
		utils.Err(c, codeInAmountFormatErr, err.Error())
		return
	}
	//out amount
	symbolDecimals := decimalsMap[req.Symbol]
	outAmount := realSwapRateDeci.Mul(inAmountDeci).Div(decimal.NewFromBigInt(big.NewInt(1), symbolDecimals-6))
	if outAmount.Cmp(swapMaxLimitDeci) > 0 {
		outAmount = swapMaxLimitDeci
	}
	if outAmount.Cmp(swapMinLimitDeci) < 0 {
		utils.Err(c, codeMinLimitErr, "out amount less than min limit")
		return
	}

	//check min out amount
	minOutAmountDeci, err := decimal.NewFromString(req.MinOutAmount)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,minOutAmount: %s err %s", req.MinOutAmount, err)
		utils.Err(c, codeMinOutAmountFormatErr, err.Error())
		return
	}
	if outAmount.Cmp(minOutAmountDeci) < 0 {
		utils.Err(c, codePriceSlideErr, "real out amount < min out amount")
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
		utils.Err(c, codeInternalErr, err.Error())
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
		utils.Err(c, codeSymbolErr, "symbol unsupport")
		return
	}
	if _, err := hexutil.Decode(blockHash); err != nil {
		utils.Err(c, codeBlockHashErr, "blockHash format err")
		return
	}
	if _, err := hexutil.Decode(txHash); err != nil {
		utils.Err(c, codeBlockHashErr, "txHash format err")
		return
	}

	swapInfo, err := dao_station.GetSwapInfoBySymbolBlkTx(h.db, symbol, strings.ToLower(blockHash), strings.ToLower(txHash))
	if err != nil && err != gorm.ErrRecordNotFound {
		utils.Err(c, codeInternalErr, err.Error())
		return
	}
	if err != nil && err == gorm.ErrRecordNotFound {
		utils.Err(c, codeSwapInfoNotExistErr, err.Error())
		return
	}

	rsp := RspSwapInfo{
		SwapStatus: swapInfo.State,
	}
	utils.Ok(c, "success", rsp)
}
