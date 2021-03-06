package cron

import (
	"time"

	"github.com/Major818/n9e/v4/src/models"
)

const cleanerInterval = 3600 * time.Second

func CleanerLoop() {
	tc := time.Tick(cleanerInterval)

	for {
		models.AuthState{}.CleanUp()
		models.Captcha{}.CleanUp()
		<-tc
	}
}
