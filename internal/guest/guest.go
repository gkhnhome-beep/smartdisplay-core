package guest

import "errors"

const (
	IDLE      = "IDLE"
	REQUESTED = "REQUESTED"
	APPROVED  = "APPROVED"
	DENIED    = "DENIED"
	EXPIRED   = "EXPIRED"
)

const (
	REQUEST = "REQUEST"
	APPROVE = "APPROVE"
	DENY    = "DENY"
	TIMEOUT = "TIMEOUT"
	EXIT    = "EXIT"
)

type StateMachine struct {
	currentState string
	lastEvent    string
}

func NewStateMachine() *StateMachine {
	return &StateMachine{currentState: IDLE}
}

func (sm *StateMachine) CurrentState() string {
	return sm.currentState
}

func (sm *StateMachine) Handle(event string) error {
	switch sm.currentState {
	case IDLE:
		switch event {
		case REQUEST:
			sm.currentState = REQUESTED
			sm.lastEvent = event
			return nil
		}
	case REQUESTED:
		switch event {
		case APPROVE:
			sm.currentState = APPROVED
			sm.lastEvent = event
			return nil
		case DENY:
			sm.currentState = DENIED
			sm.lastEvent = event
			return nil
		case TIMEOUT:
			sm.currentState = EXPIRED
			sm.lastEvent = event
			return nil
		}
	case APPROVED:
		switch event {
		case EXIT:
			sm.currentState = IDLE
			sm.lastEvent = event
			return nil
		case TIMEOUT:
			sm.currentState = EXPIRED
			sm.lastEvent = event
			return nil
		}
	case DENIED:
		switch event {
		case EXIT:
			sm.currentState = IDLE
			sm.lastEvent = event
			return nil
		}
	case EXPIRED:
		switch event {
		case EXIT:
			sm.currentState = IDLE
			sm.lastEvent = event
			return nil
		}
	}
	return errors.New("invalid transition")
}
