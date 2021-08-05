// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package station_handlers

import (
	"fee-station/pkg/db"
)

type Handler struct {
	db *db.WrapDb
}

func NewHandler(db *db.WrapDb) *Handler {
	return &Handler{db: db}
}
