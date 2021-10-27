package utils

import "github.com/shopspring/decimal"

// 0 verify sigs 1 verify tx ok 2 verify tx failed 3 swap ok
const (
	SwapStateVerifySigs        = uint8(0)
	SwapStateVerifyTxOk        = uint8(1)
	SwapStatePayOk             = uint8(2)
	SwapStateBlockHashFailed   = uint8(3)
	SwapStateTxHashFailed      = uint8(4)
	SwapStateAmountFailed      = uint8(5)
	SwapStatePubkeyFailed      = uint8(6)
	SwapStatePoolAddressFailed = uint8(7)
	SwapStateMemoFailed        = uint8(8)
)

var DecimalsMap = map[string]int32{
	SymbolAtom: 6,
	SymbolDot:  10,
	SymbolKsm:  12,
	SymbolEth:  18,
}
var DefaultSwapMaxLimitDeci = decimal.New(100, 12) //default 100e12
var DefaultSwapMinLimitDeci = decimal.New(1, 12)   //default 1e12
var DefaultSwapRateDeci = decimal.New(1, 6)        //default 1e6
