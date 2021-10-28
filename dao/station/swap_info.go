package dao_station

import "fee-station/pkg/db"

// swap info
type FeeStationSwapInfo struct {
	db.BaseModel
	StafiAddress    string `gorm:"type:varchar(80);not null;default:'0x';column:stafi_address"` //hex
	State           uint8  `gorm:"type:tinyint(1);unsigned;not null;default:0;column:state"`    //0 verify sigs 1 verify tx ok 2 verify tx failed 3 swap ok
	Symbol          string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol"`
	Blockhash       string `gorm:"type:varchar(80);not null;default:'0x';column:block_hash"`
	Txhash          string `gorm:"type:varchar(80);not null;default:'0x';column:tx_hash;uniqueIndex:uni_idx_tx"`
	PoolAddress     string `gorm:"type:varchar(80);not null;default:'';column:pool_address"`
	Signature       string `gorm:"type:varchar(150);not null;default:'0x';column:signature"`
	Pubkey          string `gorm:"type:varchar(560);not null;default:'0x';column:pubkey"` // //eth:address other:pubkey
	InAmount        string `gorm:"type:varchar(30);not null;default:'0';column:in_amount"`
	MinOutAmount    string `gorm:"type:varchar(30);not null;default:'0';column:min_out_amount"`
	OutAmount       string `gorm:"type:varchar(30);not null;default:'0';column:out_amount"`
	SwapRate        string `gorm:"type:varchar(30);not null;default:'0';column:swap_rate"` // decimal 18
	InTokenPrice    string `gorm:"type:varchar(30);not null;default:'0';column:in_token_price"`
	OutTokenPrice   string `gorm:"type:varchar(30);not null;default:'0';column:out_token_price"`
	PayInfo         string `gorm:"type:varchar(80);not null;default:'';column:pay_info"` //pay tx hash
	BundleAddressId int64  `gorm:"not null;default:0;column:bundle_address_id"`
}

func (f FeeStationSwapInfo) TableName() string {
	return "swap_infos"
}

func UpOrInFeeStationSwapInfo(db *db.WrapDb, c *FeeStationSwapInfo) error {
	return db.Save(c).Error
}

func GetFeeStationSwapInfoBySymbolBlkTx(db *db.WrapDb, symbol, blk, tx string) (info *FeeStationSwapInfo, err error) {
	info = &FeeStationSwapInfo{}
	err = db.Take(info, "symbol = ? and block_hash = ? and tx_hash = ?", symbol, blk, tx).Error
	return
}

func GetFeeStationSwapInfoByTx(db *db.WrapDb, tx string) (info *FeeStationSwapInfo, err error) {
	info = &FeeStationSwapInfo{}
	err = db.Take(info, "tx_hash = ?", tx).Error
	return
}

func GetFeeStationSwapInfoListBySymbolState(db *db.WrapDb, symbol string, state uint8) (infos []*FeeStationSwapInfo, err error) {
	err = db.Find(&infos, "symbol = ? and state = ?", symbol, state).Error
	return
}

func GetFeeStationSwapInfoListByState(db *db.WrapDb, state uint8) (infos []*FeeStationSwapInfo, err error) {
	err = db.Limit(200).Find(&infos, "state = ?", state).Error
	return
}
