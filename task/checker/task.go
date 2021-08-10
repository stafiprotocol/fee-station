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
	BlockConfirmNumber = int64(6)
)

type Task struct {
	taskTicker    int64
	atomDenom     string
	dotTypesPath  string
	ksmTypesPath  string
	coinMarketApi string
	coinGeckoApi  string
	stop          chan struct{}
	endPoint      config.Endpoint
	db            *db.WrapDb
}

func NewTask(cfg *config.Config, dao *db.WrapDb) *Task {
	s := &Task{
		taskTicker:    cfg.TaskTicker,
		atomDenom:     cfg.AtomDenom,
		dotTypesPath:  cfg.DotTypesPath,
		ksmTypesPath:  cfg.KsmTypesPath,
		coinMarketApi: cfg.CoinMarketApi,
		coinGeckoApi:  cfg.CoinGeckoApi,
		stop:          make(chan struct{}),
		endPoint:      cfg.Endpoint,
		db:            dao,
	}
	return s
}

func (task *Task) Start() error {
	utils.SafeGoWithRestart(task.AtomHandler)
	utils.SafeGoWithRestart(task.DotHandler)
	utils.SafeGoWithRestart(task.KsmHandler)
	utils.SafeGoWithRestart(task.EthHandler)
	utils.SafeGoWithRestart(task.PriceUpdateHandler)
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
			logrus.Infof("task CheckAtomTx start -----------")
			err := CheckAtomTx(task.db, task.atomDenom, task.endPoint.Atom)
			if err != nil {
				logrus.Errorf("task.CheckAtomTx err %s", err)
				utils.ShutdownRequestChannel <- struct{}{}
			}
			logrus.Infof("task CheckAtomTx end -----------")
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
			logrus.Infof("task CheckDotTx start -----------")
			err := CheckDotTx(task.db, task.endPoint.Dot, task.dotTypesPath)
			if err != nil {
				logrus.Errorf("task.CheckDotTx err %s", err)
				utils.ShutdownRequestChannel <- struct{}{}
			}
			logrus.Infof("task CheckDotTx end -----------")
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
			logrus.Infof("task CheckKsmTx start -----------")
			err := CheckKsmTx(task.db, task.endPoint.Ksm, task.ksmTypesPath)
			if err != nil {
				logrus.Errorf("task.CheckKsmTx err %s", err)
				utils.ShutdownRequestChannel <- struct{}{}
			}
			logrus.Infof("task CheckKsmTx end -----------")
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
			logrus.Infof("task CheckEthTx start -----------")
			err := CheckEthTx(task.db, task.endPoint.Eth)
			if err != nil {
				logrus.Errorf("task.CheckEthTx err %s", err)
				utils.ShutdownRequestChannel <- struct{}{}
			}
			logrus.Infof("task CheckEthTx end -----------")
		}
	}
}

func (task *Task) PriceUpdateHandler() {
	ticker := time.NewTicker(time.Duration(task.taskTicker) * time.Second)
	defer ticker.Stop()
out:
	for {
		select {
		case <-task.stop:
			logrus.Info("task has stopped")
			break out
		case <-ticker.C:

			logrus.Infof("task UpdatePrice start -----------")
			err := UpdatePrice(task.db, task.coinMarketApi, task.coinGeckoApi)
			if err != nil {
				logrus.Errorf("task.UpdatePrice err %s", err)
				utils.ShutdownRequestChannel <- struct{}{}
			}
			logrus.Infof("task UpdatePrice end -----------")
		}
	}
}
