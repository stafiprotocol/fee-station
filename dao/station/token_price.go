package dao_station

import "fee-station/pkg/db"

// token price
type TokenPrice struct {
	db.BaseModel
	Symbol string `gorm:"type:varchar(10);not null;default:'symbol';column:symbol;uniqueIndex"`
	Price  string `gorm:"type:varchar(30);not null;default:'0';column:price"` //decimals 18
}

func UpOrInTokenPrice(db *db.WrapDb, c *TokenPrice) error {
	return db.Save(c).Error
}

func GetTokenPriceBySymbol(db *db.WrapDb, symbol string) (c *TokenPrice, err error) {
	c = &TokenPrice{}
	err = db.Take(c, "symbol = ?", symbol).Error
	return
}
