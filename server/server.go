// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package server

import (
	"fee-station/api"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	listenAddr   string
	httpServer   *http.Server
	taskTicker   int64
	swapRate     string //decimals 6
	swapMaxLimit string //decimals 12
	swapMinLimit string //decimals 12
	atomDenom    string
	dotTypesPath string
	ksmTypesPath string
	poolAddress  config.PoolAddress
	endPoint     config.Endpoint
	db           *db.WrapDb
}

func NewServer(cfg *config.Config, dao *db.WrapDb) (*Server, error) {
	s := &Server{
		listenAddr:   cfg.ListenAddr,
		taskTicker:   cfg.TaskTicker,
		swapRate:     cfg.SwapRate,
		swapMaxLimit: cfg.SwapMaxLimit,
		swapMinLimit: cfg.SwapMinLimit,
		atomDenom:    cfg.AtomDenom,
		dotTypesPath: cfg.DotTypesPath,
		ksmTypesPath: cfg.KsmTypesPath,
		poolAddress:  cfg.PoolAddress,
		endPoint:     cfg.Endpoint,
		db:           dao,
	}

	cache := map[string]string{
		utils.SwapRateKey:     s.swapRate,
		utils.SwapMaxLimitKey: s.swapMaxLimit,
		utils.SwapMinLimitKey: s.swapMinLimit}

	handler := s.InitHandler(cache)

	s.httpServer = &http.Server{
		Addr:         s.listenAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return s, nil
}

func (svr *Server) InitHandler(cache map[string]string) http.Handler {
	return api.InitRouters(svr.db, cache)
}

func (svr *Server) ApiServer() {
	logrus.Infof("Gin server start on %s", svr.listenAddr)
	err := svr.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logrus.Errorf("Gin server start err: %s", err.Error())
		utils.ShutdownRequestChannel <- struct{}{} //shutdown server
		return
	}
	logrus.Infof("Gin server done on %s", svr.listenAddr)
}

//check and init dropFlowLatestDate LedgerLatestDate
func (svr *Server) InitOrUpdatePoolAddress() error {

	atom, _ := dao_station.GetFeeStationPoolAddressBySymbol(svr.db, utils.SymbolAtom)
	atom.Symbol = utils.SymbolAtom
	atom.PoolAddress = svr.poolAddress.Atom
	err := dao_station.UpOrInFeeStationPoolAddress(svr.db, atom)
	if err != nil {
		return err
	}
	eth, _ := dao_station.GetFeeStationPoolAddressBySymbol(svr.db, utils.SymbolEth)
	eth.Symbol = utils.SymbolEth
	eth.PoolAddress = svr.poolAddress.Eth
	err = dao_station.UpOrInFeeStationPoolAddress(svr.db, eth)
	if err != nil {
		return err
	}

	dot, _ := dao_station.GetFeeStationPoolAddressBySymbol(svr.db, utils.SymbolDot)
	dot.Symbol = utils.SymbolDot
	dot.PoolAddress = svr.poolAddress.Dot
	err = dao_station.UpOrInFeeStationPoolAddress(svr.db, dot)
	if err != nil {
		return err
	}

	ksm, _ := dao_station.GetFeeStationPoolAddressBySymbol(svr.db, utils.SymbolKsm)
	ksm.Symbol = utils.SymbolKsm
	ksm.PoolAddress = svr.poolAddress.Ksm
	err = dao_station.UpOrInFeeStationPoolAddress(svr.db, ksm)
	if err != nil {
		return err
	}

	return nil
}

func (svr *Server) Start() error {
	err := svr.InitOrUpdatePoolAddress()
	if err != nil {
		return err
	}
	utils.SafeGoWithRestart(svr.ApiServer)
	return nil
}

func (svr *Server) Stop() {
	if svr.httpServer != nil {
		err := svr.httpServer.Close()
		if err != nil {
			logrus.Errorf("Problem shutdown Gin server :%s", err.Error())
		}
	}
}
