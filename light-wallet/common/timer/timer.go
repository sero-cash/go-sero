package timer

import (
	"time"
	"github.com/sero-cash/go-sero/light-wallet/common/logex"
	"github.com/sero-cash/go-sero/light-wallet/common/utils"
)

type Timer struct {
	TId    string
	TName  string
	Second time.Duration
	Run    Run
}

type Run func()

func (t *Timer) Start() {
	t.TId = utils.UUID()
	timer := time.NewTicker(t.Second * time.Second)
	for {
		select {
		case <-timer.C:
			logex.Infof("Timer id=[%s], name=[%s] Run", t.TId, t.TName)
			t.Run()
			logex.Infof("Timer id=[%s], name=[%s] end, waiting %dS...", t.TId, t.TName, t.Second)
		}
	}
}
