// Copyright 2020 tpkeeper
// SPDX-License-Identifier: LGPL-3.0-only

package server

import (
	"time"
)

func (svr *Server) Task() {
	ticker := time.NewTicker(time.Duration(svr.taskTicker) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:

		}

	}
}
