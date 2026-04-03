package cronjobs

import (
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/pocketbase/pocketbase/tools/cron"
)

// Keep in mind that the app.Cron() is also used for running the system scheduled jobs like the logs cleanup or auto backups (the jobs id is in the format __pb*__) and replacing these system jobs or calling RemoveAll()/Stop() could have unintended side-effects.

// If you want more advanced control you can initialize your own cron instance independent from the application via cron.New().
var (
	_register *Register
	once      sync.Once
)

type Register struct {
	*cron.Cron
}

func NewRegister() *Register {
	once.Do(func() {
		_register = &Register{
			Cron: cron.New(),
		}
	})
	return _register
}

type Cronjob interface {
	Run()
}

type cronjobOptions struct {
	Name              string
	CronjobExpression string `env:CRONJOB_EXPRESSION`
}

func NewCronjobOptions() cronjobOptions {
	var opt cronjobOptions
	if err := env.Parse(&opt); err != nil {
		panic(err)
	}
	return opt
}
