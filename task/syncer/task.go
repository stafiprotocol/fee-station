package task

import (
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"time"

	"github.com/sirupsen/logrus"
)

// Frequency of polling for a new block
const (
	BlockRetryInterval = time.Second * 6
	BlockRetryLimit    = 100
	BlockConfirmNumber = int64(1)
)

type Task struct {
	taskTicker     int64
	atomDenom      string
	stop           chan struct{}
	syncTxEndPoint config.SyncTxEndpoint
	db             *db.WrapDb
}

func NewTask(cfg *config.Config, dao *db.WrapDb) *Task {
	s := &Task{
		taskTicker:     cfg.TaskTicker,
		atomDenom:      cfg.AtomDenom,
		stop:           make(chan struct{}),
		syncTxEndPoint: cfg.SyncTxEndpoint,
		db:             dao,
	}
	return s
}

func (task *Task) Start() error {
	utils.SafeGoWithRestart(task.AtomHandler)
	utils.SafeGoWithRestart(task.DotHandler)
	utils.SafeGoWithRestart(task.KsmHandler)
	utils.SafeGoWithRestart(task.EthHandler)
	return nil
}

func (task *Task) Stop() {
	close(task.stop)
}

func (task *Task) AtomHandler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
out:
	for {
		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:
			logrus.Infof("task syncAtomTx start -----------")
			err := SyncAtomTx(task.db, task.atomDenom, task.syncTxEndPoint.Atom)
			if err != nil {
				logrus.Errorf("task.SyncAtomTx err %s", err)
				utils.ShutdownRequestChannel <- struct{}{}
			}
			logrus.Infof("task syncAtomTx end -----------")
		}
	}
}

func (task *Task) DotHandler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
out:
	for {
		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:
			logrus.Infof("task SyncDotTx start -----------")
			err := SyncDotTx(task.db, task.syncTxEndPoint.Dot)
			if err != nil {
				logrus.Errorf("task.SyncDotTx err %s", err)
				utils.ShutdownRequestChannel <- struct{}{}
			}
			logrus.Infof("task SyncDotTx end -----------")
		}
	}
}

func (task *Task) KsmHandler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
out:
	for {
		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:
			logrus.Infof("task SyncKsmTx start -----------")
			err := SyncKsmTx(task.db, task.syncTxEndPoint.Ksm)
			if err != nil {
				logrus.Errorf("task.SyncKsmTx err %s", err)
				utils.ShutdownRequestChannel <- struct{}{}
			}
			logrus.Infof("task SyncKsmTx end -----------")
		}
	}
}
func (task *Task) EthHandler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
out:
	for {
		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:
			logrus.Infof("task SyncEthTx start -----------")
			err := SyncEthTx(task.db, task.syncTxEndPoint.Eth)
			if err != nil {
				logrus.Errorf("task.SyncEthTx err %s", err)
				utils.ShutdownRequestChannel <- struct{}{}
			}
			logrus.Infof("task SyncEthTx end -----------")
		}
	}
}
