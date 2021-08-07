package utils

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
