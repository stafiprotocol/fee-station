package utils

// 0 verify sigs 1 verify tx ok 2 verify tx failed 3 swap ok
const (
	SwapStateVerifySigs     = uint8(0)
	SwapStateVerifyTxOk     = uint8(1)
	SwapStateVerifyTxFailed = uint8(2)
	SwapStatePayOk          = uint8(3)
)
