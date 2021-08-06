// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package server

import (
	"fee-station/api"
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	listenAddr string
	httpServer *http.Server
	taskTicker int64
	swapRate   string //decimal 18
	atomDenom  string
	endPoint   config.Endpoint
	db         *db.WrapDb
}

func NewServer(cfg *config.Config, dao *db.WrapDb) (*Server, error) {
	s := &Server{
		listenAddr: cfg.ListenAddr,
		taskTicker: cfg.TaskTicker,
		swapRate:   cfg.SwapRate,
		atomDenom:  cfg.AtomDenom,
		endPoint:   cfg.Endpoint,
		db:         dao,
	}

	handler := s.InitHandler(map[string]string{
		utils.SwapRateKey: s.swapRate})

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
		shutdownRequestChannel <- struct{}{} //shutdown server
		return
	}
	logrus.Infof("Gin server done on %s", svr.listenAddr)
}

//check and init dropFlowLatestDate LedgerLatestDate
func (svr *Server) InitOrUpdateMetaData() error {
	return nil
}

func (svr *Server) Start() error {
	err := svr.InitOrUpdateMetaData()
	if err != nil {
		return err
	}
	utils.SafeGoWithRestart(svr.ApiServer)
	utils.SafeGoWithRestart(svr.Task)
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
