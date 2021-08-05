package dao_station

import "fee-station/pkg/db"

type PoolAddress struct {
	db.BaseModel
	Symbol      string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol;uniqueIndex"`
	PoolAddress string `gorm:"type:varchar(80);not null;default:'';column:pool_address"`
}

func UpOrInPoolAddress(db *db.WrapDb, c *PoolAddress) error {
	return db.Save(c).Error
}

func GetPoolAddressBySymbol(db *db.WrapDb, symbol string) (c *PoolAddress, err error) {
	c = &PoolAddress{}
	err = db.Take(c, "symbol = ?", symbol).Error
	return
}

func GetPoolAddressList(db *db.WrapDb) (list []*PoolAddress, err error) {
	err = db.Find(&list).Error
	return
}
