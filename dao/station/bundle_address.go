package dao_station

import "fee-station/pkg/db"

type FeeStationBundleAddress struct {
	db.BaseModel
	Symbol        string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol"`
	StafiAddress  string `gorm:"type:varchar(80);not null;default:'';column:stafi_address"`
	SenderAddress string `gorm:"type:varchar(80);not null;default:'';column:sender_address"` //maybe polkadot/kusama address
}

func InsertFeeStationBundleAddress(db *db.WrapDb, c *FeeStationBundleAddress) error {
	return db.Save(c).Error
}

func GetFeeStationBundleAddressListBySenderSymbol(db *db.WrapDb, sender, symbol string) (c []*FeeStationBundleAddress, err error) {
	err = db.Find(&c, "sender_address = ? and symbol = ?", sender, symbol).Error
	return
}

func GetFeeStationBundleAddressById(db *db.WrapDb, id int64) (c *FeeStationBundleAddress, err error) {
	c = &FeeStationBundleAddress{}
	err = db.Take(c, "id = ?", id).Error
	return
}
