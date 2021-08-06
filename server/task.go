// Copyright 2020 tpkeeper
// SPDX-License-Identifier: LGPL-3.0-only

package server

import (
	"fee-station/task"
	"github.com/sirupsen/logrus"
	"time"
)

func (svr *Server) Task() {
	ticker := time.NewTicker(time.Duration(svr.taskTicker) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logrus.Infof("task CheckAtomTx start -----------")
			err := task.CheckAtomTx(svr.db, svr.atomDenom, svr.endPoint.Atom)
			if err != nil {
				logrus.Errorf("task.CheckAtomTx err %s", err)
			}
			logrus.Infof("task CheckAtomTx end -----------")
		}

	}
}
