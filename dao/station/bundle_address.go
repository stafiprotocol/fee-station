package dao_station

import "fee-station/pkg/db"

type FeeStationBundleAddress struct {
	db.BaseModel
	Symbol       string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol"`
	StafiAddress string `gorm:"type:varchar(80);not null;default:'0x';column:stafi_address"` //hex
	Pubkey       string `gorm:"type:varchar(80);not null;default:'0x';column:pubkey"`        //hex maybe polkadot/kusama pubkey
	SwapRate     string `gorm:"type:varchar(30);not null;default:'0';column:swap_rate"`      // decimal 18
	PoolAddress  string `gorm:"type:varchar(80);not null;default:'';column:pool_address"`
	Signature    string `gorm:"type:varchar(150);not null;default:'0x';column:signature"`
}

func InsertFeeStationBundleAddress(db *db.WrapDb, c *FeeStationBundleAddress) error {
	return db.Create(c).Error
}

func GetFeeStationBundleAddressListByPubkeySymbol(db *db.WrapDb, pubkey, symbol string) (c []*FeeStationBundleAddress, err error) {
	err = db.Find(&c, "pubkey = ? and symbol = ?", pubkey, symbol).Error
	return
}

func GetFeeStationBundleAddressById(db *db.WrapDb, id int64) (c *FeeStationBundleAddress, err error) {
	c = &FeeStationBundleAddress{}
	err = db.Take(c, "id = ?", id).Error
	return
}
