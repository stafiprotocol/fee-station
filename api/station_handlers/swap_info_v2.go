package station_handlers

import (
	"encoding/json"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/utils"
	"math/big"
	"strings"
	"time"

	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ReqSwapInfoV2 struct {
	StafiAddress    string `json:"stafiAddress"` //hex
	Symbol          string `json:"symbol"`
	Blockhash       string `json:"blockHash"` //hex
	Txhash          string `json:"txHash"`    //hex
	PoolAddress     string `json:"poolAddress"`
	Signature       string `json:"signature"`    //hex
	Pubkey          string `json:"pubkey"`       //hex format eth:address other:pubkey
	InAmount        string `json:"inAmount"`     //decimal
	MinOutAmount    string `json:"minOutAmount"` //decimal
	BundleAddressId int64  `json:"bundleAddressId"`
}

// @Summary post swap info v2
// @Description post swap info v2
// @Tags v2
// @Accept json
// @Produce json
// @Param param body ReqSwapInfoV2 true "user swap info v2"
// @Success 200 {object} utils.Rsp{}
// @Router /v2/station/swapInfo [post]
func (h *Handler) HandlePostSwapInfoV2(c *gin.Context) {
	req := ReqSwapInfoV2{}
	err := c.Bind(&req)
	if err != nil {
		utils.Err(c, codeParamParseErr, err.Error())
		logrus.Errorf("bind err %v", err)
		return
	}
	reqBytes, _ := json.Marshal(req)
	logrus.Infof("req parm:\n %s", string(reqBytes))

	//check symbol
	if !utils.SymbolValid(req.Symbol) {
		utils.Err(c, codeSymbolErr, "symbol unsupport")
		logrus.Errorf("symbol unsupport: %s", req.Symbol)
		return
	}
	//check 0x prefix param
	var stafiAddressBytes []byte
	if stafiAddressBytes, err = hexutil.Decode(req.StafiAddress); err != nil {
		utils.Err(c, codeStafiAddressErr, "stafiAddress format err")
		logrus.Errorf("stafiAddress format err: %s", err)
		return
	}
	if len(stafiAddressBytes) != 32 {
		utils.Err(c, codeStafiAddressErr, "stafiAddress err")
		logrus.Errorf("stafiAddress len err")
		return
	}
	stafiAddressStr, err := ss58.EncodeByPubHex(req.StafiAddress[2:], ss58.StafiPrefix)
	if err != nil {
		utils.Err(c, codeStafiAddressErr, err.Error())
		logrus.Errorf("stafiAddress err %s", err)
		return
	}
	if _, err := hexutil.Decode(req.Blockhash); err != nil {
		utils.Err(c, codeBlockHashErr, "blockHash format err")
		logrus.Errorf("blockHash format err: %s", err)
		return
	}
	if _, err := hexutil.Decode(req.Txhash); err != nil {
		utils.Err(c, codeTxHashErr, "txHash format err")
		logrus.Errorf("txHash format err: %s", err)
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
	poolAddr, err := dao_station.GetFeeStationPoolAddressBySymbol(h.db, req.Symbol)
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
	swapInfo, err := dao_station.GetFeeStationSwapInfoBySymbolBlkTx(
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
		ok := utils.VerifySigsEthPersonal(sigBytes, stafiAddressStr, common.BytesToAddress(pubkeyBytes))
		if !ok {
			utils.Err(c, codeSignatureErr, "signature not right")
			logrus.Errorf("utils.VerifySigsEth failed, stafi address: %s", req.StafiAddress)
			return
		}
	}
	//get fis price
	fisPrice, err := dao_station.GetFeeStationTokenPriceBySymbol(h.db, utils.SymbolFis)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		logrus.Errorf("GetTokenPriceBySymbol %s err: %s", utils.SymbolFis, err)
		return
	}
	//check old price
	duration := int(time.Now().Unix()) - fisPrice.UpdatedAt
	if duration > priceExpiredSeconds {
		utils.Err(c, codeTokenPriceErr, "price too old")
		logrus.Errorf("fis price too old")
		return
	}

	fisPriceDeci, err := decimal.NewFromString(fisPrice.Price)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		logrus.Errorf("decimal.NewFromString(fisPrice.Price) err: %s", err)
		return
	}
	//get symbol price
	symbolPrice, err := dao_station.GetFeeStationTokenPriceBySymbol(h.db, req.Symbol)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		logrus.Errorf("GetTokenPriceBySymbol %s err: %s", req.Symbol, err)
		return
	}
	symbolPriceDeci, err := decimal.NewFromString(symbolPrice.Price)
	if err != nil {
		utils.Err(c, codeTokenPriceErr, err.Error())
		logrus.Errorf("decimal.NewFromString(symbolPrice.Price) err: %s", err)
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
		utils.Err(c, codeMinOutAmountFormatErr, err.Error())
		logrus.Errorf("decimal.NewFromString,minOutAmount: %s err %s", req.MinOutAmount, err)
		return
	}
	if outAmount.Cmp(minOutAmountDeci) < 0 {
		utils.Err(c, codePriceSlideErr, "real out amount < min out amount")
		logrus.Errorf("real out amount %s < min out amount %s", outAmount.String(), minOutAmountDeci.String())
		return
	}

	//check bundleAddressId
	_, err = dao_station.GetFeeStationBundleAddressById(h.db, req.BundleAddressId)
	if err != nil {
		utils.Err(c, codeBundleIdNotExistErr, "bundle id not exist")
		logrus.Errorf("bundle id not exit")
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
	swapInfo.BundleAddressId = req.BundleAddressId
	//update db
	err = dao_station.UpOrInFeeStationSwapInfo(h.db, swapInfo)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("UpOrInSwapInfo err %v", err)
		return
	}

	utils.Ok(c, "success", struct{}{})
}
