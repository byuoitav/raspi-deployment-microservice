package helpers

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fatih/color"
)

const STAGE_DEPLOYMENT_HOUR = 0
const PROD_DEPLOYMENT_HOUR = 0
const ACCURACY = time.Minute

var scheduledDeployments = make(map[string]bool)

func HoldDeployment(branch string, status bool) {
	if status {
		log.Printf("Disabling %s deployments", branch)
	} else {
		log.Printf("Enabling %s deployments", branch)
	}

	scheduledDeployments[branch] = status
}

// waits for s to return, then starts a deployment
func DeployOnSchedule(s chan time.Time, deviceClass, deploymentType string) {
}

// schedules a timer that returns when the specified time is reached
// accuracy specifies how to round the time
func Schedule(t time.Time, accuracy time.Duration) (chan time.Time, error) {
	// verify t is in the future
	if !t.After(time.Now()) {
		log.Printf("Can't create a schedule for a time in the past. (%s)", t)
		return nil, errors.New(fmt.Sprintf("Can't create a schedule for a time in the past. (%s)", t))
	}
	color.Set(color.FgMagenta)
	log.Printf("Creating schedule for time: %s", t)
	color.Unset()

	ret := make(chan time.Time)

	rounded := t.UTC().Round(accuracy)
	ticker := time.NewTicker(time.Nanosecond)
	go func() {
		for tick := range ticker.C {
			rtick := tick.UTC().Round(accuracy)
			//			fmt.Printf("Rounded tick: %s; compared to scheduled time: %s\n", rtick, rounded)
			if rounded.Equal(rtick) {
				color.Set(color.FgHiGreen)
				log.Printf("Scheduled time %s reached\n", rounded)
				color.Unset()
				ret <- rtick
				break
			}
		}

		ticker.Stop()
	}()
	return ret, nil
}

// returns a time tomorrow with the specified hour
func GetTimeTomorrowByHour(hour int) time.Time {
	t := time.Now()
	t = t.AddDate(0, 0, 1)
	t = time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
	return t
}
