package dao_station

import "fee-station/pkg/db"

type FeeStationPoolAddress struct {
	db.BaseModel
	Symbol      string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol;uniqueIndex"`
	PoolAddress string `gorm:"type:varchar(80);not null;default:'';column:pool_address"`
}

func (f FeeStationPoolAddress) TableName() string {
	return "pool_addresses"
}

func UpOrInFeeStationPoolAddress(db *db.WrapDb, c *FeeStationPoolAddress) error {
	return db.Save(c).Error
}

func GetFeeStationPoolAddressBySymbol(db *db.WrapDb, symbol string) (c *FeeStationPoolAddress, err error) {
	c = &FeeStationPoolAddress{}
	err = db.Take(c, "symbol = ?", symbol).Error
	return
}

func GetFeeStationPoolAddressList(db *db.WrapDb) (list []*FeeStationPoolAddress, err error) {
	err = db.Find(&list).Error
	return
}
