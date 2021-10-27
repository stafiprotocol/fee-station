package station_handlers

import (
	"encoding/json"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/utils"
	"strings"
	"time"

	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type ReqBundleAddress struct {
	StafiAddress string `json:"stafiAddress"` //hex
	Symbol       string `json:"symbol"`
	PoolAddress  string `json:"poolAddress"`
	Signature    string `json:"signature"` //hex
	Pubkey       string `json:"pubkey"`    //hex
}

type RspBundleAddress struct {
	BundleAddressId int64 `json:"bundleAddressId"`
}

// @Summary bundle address
// @Description bundle stafi address
// @Tags v1
// @Accept json
// @Produce json
// @Param param body ReqBundleAddress true "bundle address"
// @Success 200 {object} utils.Rsp{data=RspBundleAddress}
// @Router /v1/station/bundleAddress [post]
func (h *Handler) HandlePostBundleAddress(c *gin.Context) {
	req := ReqBundleAddress{}
	err := c.Bind(&req)
	if err != nil {
		utils.Err(c, codeParamParseErr, err.Error())
		logrus.Errorf("bind err %v", err)
		return
	}
	reqBytes, _ := json.Marshal(req)
	logrus.Infof("HandlePostBundleAddress req parm:\n %s", string(reqBytes))

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
		logrus.Errorf("stafiAddress format err: %s", req.StafiAddress)
		return
	}
	if len(stafiAddressBytes) != 32 {
		utils.Err(c, codeStafiAddressErr, "stafiAddress err")
		logrus.Errorf("stafiAddress len err: %s", req.StafiAddress)
		return
	}
	stafiAddressStr, err := ss58.EncodeByPubHex(req.StafiAddress[2:], ss58.StafiPrefix)
	if err != nil {
		utils.Err(c, codeStafiAddressErr, err.Error())
		logrus.Errorf("ss58.EncodeByPubHex err: %s", err)
		return
	}

	var sigBytes []byte
	var pubkeyBytes []byte
	if sigBytes, err = hexutil.Decode(req.Signature); req.Symbol != utils.SymbolAtom && err != nil {
		utils.Err(c, codeSignatureErr, "signature format err")
		logrus.Errorf("decode signature err: %s", err)
		return
	}
	if pubkeyBytes, err = hexutil.Decode(req.Pubkey); err != nil {
		utils.Err(c, codePubkeyErr, "pubkey format err")
		logrus.Errorf("decode pubkey err: %s", err)
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
		logrus.Errorf("price too old")
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
	swapRateDeci, err := decimal.NewFromString(swapRateStr)
	if err != nil {
		logrus.Errorf("decimal.NewFromString,swapRateStr: %s err %s", swapRateStr, err)
		swapRateDeci = utils.DefaultSwapRateDeci
	}

	//cal real swap rate
	realSwapRateDeci := swapRateDeci.Mul(symbolPriceDeci).Div(fisPriceDeci)

	bundleAddress := dao_station.FeeStationBundleAddress{}
	bundleAddress.StafiAddress = strings.ToLower(req.StafiAddress)
	bundleAddress.Symbol = req.Symbol
	bundleAddress.PoolAddress = req.PoolAddress
	bundleAddress.Signature = strings.ToLower(req.Signature)
	bundleAddress.Pubkey = strings.ToLower(req.Pubkey)
	bundleAddress.SwapRate = realSwapRateDeci.StringFixed(0)

	//update db
	err = dao_station.InsertFeeStationBundleAddress(h.db, &bundleAddress)
	if err != nil {
		utils.Err(c, codeInternalErr, err.Error())
		logrus.Errorf("InsertFeeStationBundleAddress err %v", err)
		return
	}
	rsp := RspBundleAddress{
		BundleAddressId: bundleAddress.ID,
	}

	utils.Ok(c, "success", rsp)
}
