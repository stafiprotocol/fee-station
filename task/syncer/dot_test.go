package task_test

import (
	task "fee-station/task/syncer"
	"testing"
)

func TestGetDot(t *testing.T) {
	tx,err:=task.GetSubstrateTxs("https://polkadot.api.subscan.io/api/scan/transfers","13frq3FZeKV8Zzzq68AEbjoDoiAr7iWYerxcG7NTBpwQ68gF","",0,100)
	if err!=nil{
		t.Fatal(err)
	}
	t.Log(tx)
}
