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
	return &Handler{db: db, cache: cache}
}

const (
	codeParamParseErr         = "80001"
	codeSymbolErr             = "80002"
	codeStafiAddressErr       = "80003"
	codeBlockHashErr          = "80004"
	codeTxHashErr             = "80005"
	codeSignatureErr          = "80006"
	codePubkeyErr             = "80007"
	codeInternalErr           = "80008"
	codePoolAddressErr        = "80009"
	codeTxDuplicateErr        = "80010"
	codeTokenPriceErr         = "80011"
	codeInAmountFormatErr     = "80012"
	codeMinOutAmountFormatErr = "80013"
	codePriceSlideErr         = "80014"
	codeMinLimitErr           = "80015"
	codeMaxLimitErr           = "80016"
	codeSwapInfoNotExistErr   = "80017"
	codeBundleIdNotExistErr   = "80018"
)
