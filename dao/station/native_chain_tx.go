package dao_station

import "fee-station/pkg/db"

// native chain transaction
type FeeStationNativeChainTx struct {
	db.BaseModel
	State         uint8  `gorm:"type:tinyint(1);unsigned;not null;default:0;column:state"` //0: not deal 1: has deal
	Symbol        string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol"`
	Blockhash     string `gorm:"type:varchar(80);not null;default:'0x';column:block_hash"`
	Txhash        string `gorm:"type:varchar(80);not null;default:'0x';column:tx_hash;uniqueIndex:uni_idx_tx"`
	PoolAddress   string `gorm:"type:varchar(80);not null;default:'';column:pool_address"`
	SenderAddress string `gorm:"type:varchar(80);not null;default:'';column:sender_address"`
	InAmount      string `gorm:"type:varchar(30);not null;default:'0';column:in_amount"`
}

func UpOrInFeeStationNativeChainTx(db *db.WrapDb, c *FeeStationNativeChainTx) error {
	return db.Save(c).Error
}

func GetFeeStationNativeChainTxByTxhash(db *db.WrapDb, symbol, tx string) (info *FeeStationNativeChainTx, err error) {
	info = &FeeStationNativeChainTx{}
	err = db.Take(info, "symbol = ? and tx_hash = ?", symbol, tx).Error
	return
}

func GetFeeStationNativeTxBySymbolState(db *db.WrapDb, symbol string, state uint8) (infos []*FeeStationNativeChainTx, err error) {
	err = db.Find(&infos, "symbol = ? and state = ?", symbol, state).Error
	return
}
