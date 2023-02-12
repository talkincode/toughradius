package events

import (
	evbus "github.com/asaskevich/EventBus"
)

var (
	Supervisor = evbus.New()
)

const (
	EventSuperviseLog        = "EventSuperviseLog"
	EventSuperviseStatus     = "EventSuperviseStatus"
	EventCwmpSuperviseStatus = "EventCwmpSuperviseStatus"
)

func PubSuperviseLog(devid int64, session, level, message string) {
	Supervisor.Publish(EventSuperviseLog, devid, session, level, message)
}

func PubSuperviseStatus(devid int64, action, message string) {
	Supervisor.Publish(EventSuperviseStatus, devid, action, message)
}

func PubEventCwmpSuperviseStatus(sn, session, level, message string) {
	Supervisor.Publish(EventCwmpSuperviseStatus, sn, session, level, message)
}
