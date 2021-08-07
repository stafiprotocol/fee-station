// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package api

import (
	"fee-station/api/station_handlers"
	"fee-station/pkg/db"
	"net/http"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

func InitRouters(db *db.WrapDb, cache map[string]string) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.Static("/static", "./static")
	router.Use(Cors())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	stationHandler := station_handlers.NewHandler(db, cache)
	router.GET("/feeStation/api/v1/station/poolInfo", stationHandler.HandleGetPoolInfo)
	router.POST("/feeStation/api/v1/station/swapInfo", stationHandler.HandlePostSwapInfo)
	router.GET("/feeStation/api/v1/station/swapInfo", stationHandler.HandleGetSwapInfo)

	return router
}
