package task

import (
	"fee-station/pkg/config"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"time"

	"github.com/sirupsen/logrus"
)

type Task struct {
	taskTicker   int64
	atomDenom    string
	dotTypesPath string
	ksmTypesPath string
	stop         chan struct{}
	endPoint     config.Endpoint
	db           *db.WrapDb
}

func NewTask(cfg *config.Config, dao *db.WrapDb) *Task {
	s := &Task{
		taskTicker:   cfg.TaskTicker,
		atomDenom:    cfg.AtomDenom,
		dotTypesPath: cfg.DotTypesPath,
		ksmTypesPath: cfg.KsmTypesPath,
		stop:         make(chan struct{}),
		endPoint:     cfg.Endpoint,
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
			logrus.Infof("task CheckAtomTx start -----------")
			err := CheckAtomTx(task.db, task.atomDenom, task.endPoint.Atom)
			if err != nil {
				logrus.Errorf("task.CheckAtomTx err %s", err)
			}
			logrus.Infof("task CheckAtomTx end -----------")
			logrus.Infof("task CheckDotTx start -----------")
			err = CheckDotTx(task.db, task.endPoint.Dot, task.dotTypesPath)
			if err != nil {
				logrus.Errorf("task.CheckDotTx err %s", err)
			}
			logrus.Infof("task CheckDotTx end -----------")

			logrus.Infof("task CheckKsmTx start -----------")
			err = CheckKsmTx(task.db, task.endPoint.Ksm, task.ksmTypesPath)
			if err != nil {
				logrus.Errorf("task.CheckKsmTx err %s", err)
			}
			logrus.Infof("task CheckKsmTx end -----------")
		}
	}
}
