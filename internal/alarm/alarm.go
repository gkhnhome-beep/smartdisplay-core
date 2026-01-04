package alarm

import (
	"errors"
	"smartdisplay-core/internal/alarm/countdown"
	"smartdisplay-core/internal/logger"
)

const (
	DISARMED  = "DISARMED"
	ARMING    = "ARMING"
	ARMED     = "ARMED"
	TRIGGERED = "TRIGGERED"
)

const (
	ARM_REQUEST    = "ARM_REQUEST"
	DISARM_REQUEST = "DISARM_REQUEST"
	ARM_COMPLETE   = "ARM_COMPLETE"
	TRIGGER        = "TRIGGER"
	RESET          = "RESET"
)

	currentState   string
	lastEvent      string
	lastTrigger    time.Time
	cd             *countdown.Countdown
	guest          interface{ CurrentState() string }
}
func (sm *StateMachine) SetGuest(guest interface{ CurrentState() string }) {
   sm.guest = guest
}

func (sm *StateMachine) CanDisarmRequest() bool {
   if sm.guest == nil {
	   logger.Info("guest not set: disarm allowed")
	   return true
   }
   switch sm.guest.CurrentState() {
   case "APPROVED":
	   logger.Info("guest approved: disarm allowed")
	   return true
   case "DENIED", "EXPIRED":
	   logger.Info("guest denied/expired: disarm not allowed")
	   return false
   default:
	   logger.Info("guest state: disarm not allowed")
	   return false
   }
}

func (sm *StateMachine) CanArmRequest() bool {
   if sm.guest == nil {
	   logger.Info("guest not set: arm allowed")
	   return true
   }
   if sm.guest.CurrentState() == "EXIT" {
	   logger.Info("guest exit: arm allowed")
	   return true
   }
   logger.Info("guest state: arm not allowed")
   return false
}

func NewStateMachine() *StateMachine {
	return &StateMachine{currentState: DISARMED}
}

func (sm *StateMachine) CurrentState() string {
	return sm.currentState
}

func (sm *StateMachine) Handle(event string) error {
   switch sm.currentState {
   case DISARMED:
	   switch event {
	   case ARM_REQUEST:
		   sm.currentState = ARMING
		   sm.lastEvent = event
		   sm.cd = countdown.New(30)
		   logger.Info("alarm countdown created")
		   return nil
	   }
   case ARMING:
	   switch event {
	   case ARM_COMPLETE:
		   sm.currentState = ARMED
		   sm.lastEvent = event
		   if sm.cd != nil {
			   sm.cd.Stop()
		   }
		   return nil
	   case DISARM_REQUEST:
		   sm.currentState = DISARMED
		   sm.lastEvent = event
		   if sm.cd != nil {
			   sm.cd.Reset()
			   logger.Info("alarm countdown reset")
		   }
		   return nil
	   }
   case ARMED:
	   switch event {
	   case DISARM_REQUEST:
		   sm.currentState = DISARMED
		   sm.lastEvent = event
		   if sm.cd != nil {
			   sm.cd.Reset()
			   logger.Info("alarm countdown reset")
		   }
		   return nil
	   case TRIGGER:
		   sm.currentState = TRIGGERED
		   sm.lastEvent = event
		   sm.lastTrigger = time.Now()
		   return nil
	   }
   case TRIGGERED:
	   switch event {
	   case RESET:
		   sm.currentState = DISARMED
		   sm.lastEvent = event
		   return nil
	   }
   }
   return errors.New("invalid transition")
}

// LastEvent returns the last event handled by the alarm state machine
func (sm *StateMachine) LastEvent() string {
	return sm.lastEvent
}

// LastTriggerTime returns the last time the alarm was triggered
func (sm *StateMachine) LastTriggerTime() time.Time {
	return sm.lastTrigger
}
