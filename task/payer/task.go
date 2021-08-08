package task

import (
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"time"

	"github.com/sirupsen/logrus"
)

type Task struct {
	taskTicker   int64
	fisTypesPath string
	keystorePath string
	fisEndpoint  string
	stop         chan struct{}
	db           *db.WrapDb
}

func NewTask(cfg *config.Config, dao *db.WrapDb) *Task {
	s := &Task{
		taskTicker:   cfg.TaskTicker,
		fisTypesPath: cfg.FisTypesPath,
		keystorePath: cfg.KeystorePath,
		fisEndpoint:  cfg.FisEndpoint,
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
		}
	}
}
