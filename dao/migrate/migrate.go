// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package migrate

import (
	"fee-station/dao/station"
	"fee-station/pkg/db"
	"fmt"
)

func AutoMigrate(db *db.WrapDb) error {
	err := dao_station.AutoMigrate(db)
	if err != nil {
		return fmt.Errorf("dao_user.AutoMigrate %s", err)
	}
	return nil
}
