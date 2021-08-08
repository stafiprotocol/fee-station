package task

import (
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"github.com/sirupsen/logrus"
	"github.com/stafiprotocol/go-substrate-rpc-client/signature"
	"time"
)

type Task struct {
	taskTicker   int64
	fisTypesPath string
	keystorePath string
	fisEndpoint  string
	payerAccount string
	swapLimit    string
	key          *signature.KeyringPair
	stop         chan struct{}
	db           *db.WrapDb
}

func NewTask(cfg *config.Config, dao *db.WrapDb, key *signature.KeyringPair) *Task {
	s := &Task{
		taskTicker:   cfg.TaskTicker,
		fisTypesPath: cfg.FisTypesPath,
		keystorePath: cfg.KeystorePath,
		fisEndpoint:  cfg.FisEndpoint,
		payerAccount: cfg.PayerAccount,
		swapLimit:    cfg.SwapLimit,
		key:          key,
		stop:         make(chan struct{}),
		db:           dao,
	}
	return s
}

func (task *Task) Start() error {
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
			err := CheckPayInfo(task.db, task.fisTypesPath, task.fisEndpoint, task.swapLimit, task.key)
			if err != nil {
				logrus.Errorf("task.CheckPayInfo err %s", err)
			}
			logrus.Infof("task CheckPayInfo end -----------")
		}
	}
}
