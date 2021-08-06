// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package station_handlers

import (
	"fee-station/pkg/db"
)

type Handler struct {
	db    *db.WrapDb
	cache map[string]string
}

func NewHandler(db *db.WrapDb, cache map[string]string) *Handler {
	return &Handler{db: db}
}
