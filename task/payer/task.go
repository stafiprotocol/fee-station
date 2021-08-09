package task

import (
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stafiprotocol/go-substrate-rpc-client/signature"
)

// Frequency of polling for a new block
const (
	BlockRetryInterval = time.Second * 6
	BlockRetryLimit    = 100
	BlockConfirmNumber = int64(6)
)

type Task struct {
	taskTicker   int64
	keystorePath string
	fisEndpoint  string
	payerAccount string
	swapMaxLimit string
	key          *signature.KeyringPair
	stop         chan struct{}
	db           *db.WrapDb
}

func NewTask(cfg *config.Config, dao *db.WrapDb, key *signature.KeyringPair) *Task {
	s := &Task{
		taskTicker:   cfg.TaskTicker,
		keystorePath: cfg.KeystorePath,
		fisEndpoint:  cfg.FisEndpoint,
		payerAccount: cfg.PayerAccount,
		swapMaxLimit: cfg.SwapMaxLimit,
		key:          key,
		stop:         make(chan struct{}),
		db:           dao,
	}
	return s
}

func (task *Task) Start() error {
	utils.SafeGoWithRestart(task.Handler)
	return nil
}

func (task *Task) Stop() {
	close(task.stop)
}

func (task *Task) Handler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
out:
	for {
		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:
			logrus.Infof("task CheckPayInfo start -----------")
			err := CheckPayInfo(task.db, task.fisEndpoint, task.swapMaxLimit, task.key)
			if err != nil {
				logrus.Errorf("task.CheckPayInfo err %s", err)
				utils.ShutdownRequestChannel <- struct{}{}
			}
			logrus.Infof("task CheckPayInfo end -----------")
		}
	}
}
