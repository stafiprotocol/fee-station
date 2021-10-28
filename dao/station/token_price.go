package dao_station

import "fee-station/pkg/db"

// token price
type FeeStationTokenPrice struct {
	db.BaseModel
	Symbol string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol;uniqueIndex"`
	Price  string `gorm:"type:varchar(30);not null;default:'0';column:price"` //decimals 18
}

func (f FeeStationTokenPrice) TableName() string {
	return "token_prices"
}

func UpOrInFeeStationTokenPrice(db *db.WrapDb, c *FeeStationTokenPrice) error {
	return db.Save(c).Error
}

func GetFeeStationTokenPriceBySymbol(db *db.WrapDb, symbol string) (c *FeeStationTokenPrice, err error) {
	c = &FeeStationTokenPrice{}
	err = db.Take(c, "symbol = ?", symbol).Error
	return
}
