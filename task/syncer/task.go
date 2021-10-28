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
	taskTicker      int64
	recoverInterval int64
	startTimestamp  int64
	swapMaxLimit    string
	atomDenom       string
	stop            chan struct{}
	etherScanApiKey string
	subScanApiKey   string
	syncTxEndPoint  config.SyncTxEndpoint
	db              *db.WrapDb
}

func NewTask(cfg *config.Config, dao *db.WrapDb) *Task {
	s := &Task{
		taskTicker:      cfg.TaskTicker,
		recoverInterval: cfg.RecoverInterval,
		startTimestamp:  cfg.StartTimestamp,
		swapMaxLimit:    cfg.SwapMaxLimit,
		atomDenom:       cfg.AtomDenom,
		stop:            make(chan struct{}),
		etherScanApiKey: cfg.EtherScanApiKey,
		subScanApiKey:   cfg.SubScanApiKey,
		syncTxEndPoint:  cfg.SyncTxEndpoint,
		db:              dao,
	}
	return s
}

func (task *Task) Start() error {
	utils.SafeGoWithRestart(task.AtomHandler)
	utils.SafeGoWithRestart(task.DotHandler)
	utils.SafeGoWithRestart(task.KsmHandler)
	utils.SafeGoWithRestart(task.EthHandler)
	utils.SafeGoWithRestart(task.RecoverHandler)
	return nil
}

func (task *Task) Stop() {
	close(task.stop)
}

func (task *Task) AtomHandler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
	retry := 0
out:
	for {

		if retry > BlockRetryLimit {
			utils.ShutdownRequestChannel <- struct{}{}
		}

		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:
			logrus.Infof("task syncAtomTx start -----------")
			err := SyncAtomTx(task.db, task.atomDenom, task.syncTxEndPoint.Atom)
			if err != nil {
				logrus.Errorf("task.SyncAtomTx err %s", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue out
			}
			logrus.Infof("task syncAtomTx end -----------")
			retry = 0
		}
	}
}

func (task *Task) DotHandler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
	retry := 0
out:
	for {
		if retry > BlockRetryLimit {
			utils.ShutdownRequestChannel <- struct{}{}
		}
		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:
			logrus.Infof("task SyncDotTx start -----------")
			err := SyncDotTx(task.db, task.syncTxEndPoint.Dot, task.subScanApiKey)
			if err != nil {
				logrus.Errorf("task.SyncDotTx err %s", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue out
			}
			logrus.Infof("task SyncDotTx end -----------")
			retry = 0
		}
	}
}

func (task *Task) KsmHandler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
	retry := 0
out:
	for {
		if retry > BlockRetryLimit {
			utils.ShutdownRequestChannel <- struct{}{}
		}
		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:
			logrus.Infof("task SyncKsmTx start -----------")
			err := SyncKsmTx(task.db, task.syncTxEndPoint.Ksm, task.subScanApiKey)
			if err != nil {
				logrus.Errorf("task.SyncKsmTx err %s", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue out
			}
			logrus.Infof("task SyncKsmTx end -----------")
			retry = 0
		}
	}
}

func (task *Task) EthHandler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
	retry := 0
out:
	for {
		if retry > BlockRetryLimit {
			utils.ShutdownRequestChannel <- struct{}{}
		}
		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:
			logrus.Infof("task SyncEthTx start -----------")
			err := SyncEthTx(task.db, task.syncTxEndPoint.Eth, task.etherScanApiKey)
			if err != nil {
				logrus.Errorf("task.SyncEthTx err %s", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue out
			}
			logrus.Infof("task SyncEthTx end -----------")
			retry = 0
		}
	}
}

func (task *Task) RecoverHandler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
	retry := 0
out:
	for {
		if retry > BlockRetryLimit {
			utils.ShutdownRequestChannel <- struct{}{}
		}
		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:
			logrus.Infof("task Recover start -----------")
			err := Recover(task.db, task.recoverInterval, task.startTimestamp, task.swapMaxLimit)
			if err != nil {
				logrus.Errorf("task.Recover err %s", err)
				time.Sleep(BlockRetryInterval)
				retry++
				continue out
			}
			logrus.Infof("task Recover end -----------")
			retry = 0
		}
	}
}
