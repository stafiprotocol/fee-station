package dao_station

import "fee-station/pkg/db"

// swap info
type SwapInfo struct {
	db.BaseModel
	StafiAddress  string `gorm:"type:varchar(42);not null;default:'';column:stafi_address"` //base58 address
	State         uint8  `gorm:"type:tinyint(1);unsigned;not null;default:0;column:state"`  //0 verify sigs 1 verify tx ok 2 verify tx failed 3 swap ok
	Symbol        string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol"`
	Blockhash     string `gorm:"type:varchar(80);not null;default:'0x';column:block_hash;uniqueIndex:uni_idx_blk_tx"`
	Txhash        string `gorm:"type:varchar(80);not null;default:'0x';column:tx_hash;uniqueIndex:uni_idx_blk_tx"`
	PoolAddress   string `gorm:"type:varchar(80);not null;default:'';column:pool_address"`
	Signature     string `gorm:"type:varchar(100);not null;default:'0x';column:signature"`
	Pubkey        string `gorm:"type:varchar(70);not null;default:'0x';column:pubkey"`
	InAmount      string `gorm:"type:varchar(30);not null;default:'0';column:in_amount"`
	MinOutAmount  string `gorm:"type:varchar(30);not null;default:'0';column:min_out_amount"`
	OutAmount     string `gorm:"type:varchar(30);not null;default:'0';column:out_amount"`
	SwapRate      string `gorm:"type:varchar(30);not null;default:'0';column:swap_rate"` // decimal 18
	InTokenPrice  string `gorm:"type:varchar(30);not null;default:'0';column:in_token_price"`
	OutTokenPrice string `gorm:"type:varchar(30);not null;default:'0';column:out_token_price"`
	PayInfo       string `gorm:"type:varchar(80);not null;default:'';column:pay_info"` //pay tx hash
}

func UpOrInSwapInfo(db *db.WrapDb, c *SwapInfo) error {
	return db.Save(c).Error
}

func GetSwapInfoBySymbolBlkTx(db *db.WrapDb, symbol, blk, tx string) (info *SwapInfo, err error) {
	info = &SwapInfo{}
	err = db.Take(info, "symbol = ? and block_hash = ? and tx_hash = ?", symbol, blk, tx).Error
	return
}

func GetSwapInfoListByState(db *db.WrapDb, state uint8) (infos []*SwapInfo, err error) {
	err = db.Find(&infos, "state = ?", state).Error
	return
}
