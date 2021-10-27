package dao_station

import "fee-station/pkg/db"

// native chain transaction
type FeeStationNativeChainTx struct {
	db.BaseModel
	State        uint8  `gorm:"type:tinyint(1);unsigned;not null;default:0;column:state"` //0: not deal 1: has deal
	TxStatus     int64  `gorm:"unsigned;not null;default:0;column:tx_status"`             //tx status 0: success 1: failed 2: not receive
	Symbol       string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol"`
	Blockhash    string `gorm:"type:varchar(80);not null;default:'0x';column:block_hash"`
	Txhash       string `gorm:"type:varchar(80);not null;default:'0x';column:tx_hash;uniqueIndex:uni_idx_tx"`
	PoolAddress  string `gorm:"type:varchar(80);not null;default:'';column:pool_address"`
	SenderPubkey string `gorm:"type:varchar(80);not null;default:'0x';column:sender_pubkey"` //eth:address other:pubkey
	InAmount     string `gorm:"type:varchar(30);not null;default:'0';column:in_amount"`
	TxTimestamp  int64  `gorm:"unsigned;not null;default:0;column:tx_timestamp"`
}

func UpOrInFeeStationNativeChainTx(db *db.WrapDb, c *FeeStationNativeChainTx) error {
	return db.Save(c).Error
}

func GetFeeStationNativeChainTxBySymbolTxhash(db *db.WrapDb, symbol, tx string) (info *FeeStationNativeChainTx, err error) {
	info = &FeeStationNativeChainTx{}
	err = db.Take(info, "symbol = ? and tx_hash = ?", symbol, tx).Error
	return
}

func GetFeeStationNativeChainTxTotalCount(db *db.WrapDb, symbol string) (count int64, err error) {
	err = db.Model(&FeeStationNativeChainTx{}).Where("symbol = ?", symbol).Count(&count).Error
	return
}

func GetFeeStationNativeTxBySymbolState(db *db.WrapDb, symbol string, state uint8) (infos []*FeeStationNativeChainTx, err error) {
	err = db.Find(&infos, "symbol = ? and state = ?", symbol, state).Error
	return
}

func GetFeeStationNativeTxByState(db *db.WrapDb, state uint8, txStatus, startTimestamp int64) (infos []*FeeStationNativeChainTx, err error) {
	err = db.Find(&infos, "state = ? and tx_status = ? and tx_timestamp > ?", state, txStatus, startTimestamp).Error
	return
}
