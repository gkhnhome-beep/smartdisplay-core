package haadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"smartdisplay-core/internal/logger"
	"strings"
	"sync"
	"time"
)

const (
	EventAlarmStateChange = "alarm_state_change"
	EventGuestRequest     = "guest_request"
	EventGuestResponse    = "guest_response"
)

type Event struct {
	Type    string
	Payload map[string]interface{}
}

type Adapter struct {
	connected bool
	mu        sync.Mutex
	baseURL   string
	token     string
	ws        interface{} // Placeholder for websocket connection
}

// Domain model references (to be set by main or test)
var alarmSM interface{ Handle(string) error }
var guestSM interface{ Handle(string) error }

func SetAlarmStateMachine(sm interface{ Handle(string) error }) {
	alarmSM = sm
}

func SetGuestStateMachine(sm interface{ Handle(string) error }) {
	guestSM = sm
}

func (a *Adapter) HandleEvent(event Event) error {
	switch event.Type {
	case EventAlarmStateChange:
		if alarmSM != nil && event.Payload["event"] != nil {
			logger.Info("ha event: alarm_state_change → alarm.Handle(" + event.Payload["event"].(string) + ")")
			return alarmSM.Handle(event.Payload["event"].(string))
		}
	case EventGuestRequest:
		if guestSM != nil {
			logger.Info("ha event: guest_request → guest.Handle(REQUEST)")
			return guestSM.Handle("REQUEST")
		}
	case EventGuestResponse:
		if guestSM != nil && event.Payload["response"] != nil {
			resp := event.Payload["response"].(string)
			if resp == "APPROVE" || resp == "DENY" {
				logger.Info("ha event: guest_response → guest.Handle(" + resp + ")")
				return guestSM.Handle(resp)
			}
		}
	}
	logger.Info("ha event: unhandled or missing payload")
	return nil
}

func New() *Adapter {
	baseURL := os.Getenv("HA_BASE_URL")
	token := os.Getenv("HA_TOKEN")
	if baseURL == "" || token == "" {
		logger.Error("HA_BASE_URL or HA_TOKEN not set in environment")
	}
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		logger.Error("invalid HA_BASE_URL format")
	}
	return &Adapter{baseURL: baseURL, token: token}
}

func (a *Adapter) Init() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connected = false
	logger.Info("ha adapter initialized")
	return nil
}

func (a *Adapter) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	// Placeholder for websocket connection (would use golang.org/x/net/websocket in full implementation)
	a.connected = true
	logger.Info("ha adapter started (placeholder)")
	return nil
}

// REST: CallService(domain, service, payload)
func (a *Adapter) CallService(domain, service string, payload map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.connected {
		return errors.New("not connected")
	}
	url := a.baseURL + "/api/services/" + domain + "/" + service
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		logger.Error("callservice request error: " + err.Error())
		return err
	}
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("callservice http error: " + err.Error())
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logger.Error("callservice failed: status " + resp.Status)
		return errors.New("callservice failed")
	}
	logger.Info("callservice success: " + domain + "." + service)
	return nil
}

// Alarm arm/disarm helpers
func (a *Adapter) ArmAlarm(payload map[string]interface{}) error {
	return a.CallService("alarm_control_panel", "alarm_arm_away", payload)
}

func (a *Adapter) DisarmAlarm(payload map[string]interface{}) error {
	return a.CallService("alarm_control_panel", "alarm_disarm", payload)
}

func (a *Adapter) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.connected = false
	logger.Info("ha adapter stopped")
}

func (a *Adapter) IsConnected() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.connected
}
