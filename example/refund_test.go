package example

import (
	"github.com/robfig/cron/v3"
	"github.com/scorpiotzh/mylog"
	"testing"
	"time"
)

var (
	log = mylog.NewLogger("example", mylog.LevelDebug)
)

func TestCron(t *testing.T) {

	c := cron.New(cron.WithSeconds())

	_, err := c.AddFunc("0 0 0 1/1 * *", func() {
		log.Info(time.Now().String())
	})
	if err != nil {
		t.Fatal(err)
	}
	c.Start()

	select {}

}
