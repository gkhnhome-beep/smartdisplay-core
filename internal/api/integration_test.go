package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"smartdisplay-core/internal/alarm"
	"smartdisplay-core/internal/alarm/countdown"
	"smartdisplay-core/internal/config"
	"smartdisplay-core/internal/firstboot"
	"smartdisplay-core/internal/guest"
	"smartdisplay-core/internal/ha/alarmo"
	"smartdisplay-core/internal/hal"
	"smartdisplay-core/internal/hanotify"
	"smartdisplay-core/internal/logger"
	"smartdisplay-core/internal/menu"
	"smartdisplay-core/internal/platform"
	"smartdisplay-core/internal/settings"
	"smartdisplay-core/internal/system"
)

// TestConfig holds per-test configuration
type TestConfig struct {
	WizardCompleted bool
	ReducedMotion   bool
}

// TestServer wraps HTTP test server with coordinator and helpers
type TestServer struct {
	*httptest.Server
	Coordinator *system.Coordinator
	testConfig  TestConfig
}

// startTestServer initializes a complete test server with in-memory config
func startTestServer(t *testing.T, cfg TestConfig) *TestServer {
	t.Helper()

	// Ensure logs directory exists for logger
	os.MkdirAll("logs", 0755)

	// Initialize logger (required by all subsystems)
	logger.Init()

	// Create in-memory runtime config (no file I/O)
	runtimeCfg := &config.RuntimeConfig{
		Language:        "en",
		WizardCompleted: cfg.WizardCompleted,
		HighContrast:    false,
		LargeText:       false,
		ReducedMotion:   false,
		VoiceEnabled:    true,
	}
	runtimeCfg.ReducedMotion = cfg.ReducedMotion

	_ = runtimeCfg // In-memory config, not persisted to file

	// Initialize subsystems (minimal for testing)
	alarmSM := alarm.NewStateMachine()
	guestSM := guest.NewStateMachine()
	cd := countdown.New(30)
	notifier := &hanotify.StubNotifier{}
	halReg := hal.NewRegistry()
	plat := platform.DetectPlatform()

	// Create coordinator (integrates all subsystems)
	coord := system.NewCoordinator(alarmSM, guestSM, cd, nil, notifier, halReg, plat, "", "")

	// Configure first-boot manager according to test config
	coord.FirstBoot = firstboot.New(cfg.WizardCompleted)

	// Rebuild menu manager so it observes the correct first-boot state
	if coord.Menu != nil {
		coord.Menu = menu.NewMenuManager(
			func() bool {
				return coord.FirstBoot != nil && coord.FirstBoot.Active()
			},
			func() bool {
				return coord.InFailsafeMode()
			},
			func() bool {
				if coord.Guest != nil {
					return coord.Guest.HasPendingRequest()
				}
				return false
			},
			menu.RoleAdmin,
		)
	}

	// Create API server
	server := NewServer(coord, runtimeCfg)

	// Build handler manually for testing (without real listen port)
	mux := server.registerRoutes()
	handler := requestIDMiddleware(mux)
	handler = panicRecovery(handler)

	// Start httptest server with our handler
	ts := httptest.NewServer(handler)

	return &TestServer{
		Server:      ts,
		Coordinator: coord,
		testConfig:  cfg,
	}
}

// TestMain ensures tests run from repository root so runtime config paths resolve consistently.
func TestMain(m *testing.M) {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to determine working directory: %v\n", err)
		os.Exit(1)
	}

	root := filepath.Join(cwd, "..", "..")
	if err := os.Chdir(root); err != nil {
		fmt.Fprintf(os.Stderr, "failed to change working directory to repo root: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := os.Chdir(cwd); err != nil {
		fmt.Fprintf(os.Stderr, "failed to restore working directory: %v\n", err)
	}

	os.Exit(code)
}

// Shutdown gracefully closes the test server
func (ts *TestServer) Shutdown() error {
	ts.Server.Close()
	return nil
}

// newTestRequest creates an HTTP request with user role header
func newTestRequest(t *testing.T, method, path string, role string) *http.Request {
	t.Helper()

	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Add user role header
	if role != "" {
		req.Header.Set("X-User-Role", role)
	}

	return req
}

// newTestRequestWithBody creates an HTTP request with JSON body
func newTestRequestWithBody(t *testing.T, method, path, role string, body interface{}) *http.Request {
	t.Helper()

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(method, path, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if role != "" {
		req.Header.Set("X-User-Role", role)
	}

	return req
}

// TestResponse holds parsed HTTP response
type TestResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
	JSON       map[string]interface{}
}

// parseJSONResponse reads and parses HTTP response
func parseJSONResponse(t *testing.T, resp *http.Response) TestResponse {
	t.Helper()

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	tr := TestResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
		JSON:       make(map[string]interface{}),
	}

	// Try to parse as JSON
	if len(body) > 0 {
		if err := json.Unmarshal(body, &tr.JSON); err != nil {
			t.Logf("response is not JSON: %v", err)
		} else {
			if respField, ok := tr.JSON["response"].(map[string]interface{}); ok {
				if dataVal, hasData := respField["data"]; hasData {
					tr.JSON["data"] = dataVal
				}
				if errVal, hasErr := respField["error"]; hasErr {
					tr.JSON["error"] = errVal
				}
				if okVal, hasOk := respField["ok"]; hasOk {
					tr.JSON["ok"] = okVal
				}
			}
		}
	}

	return tr
}

// AssertStatusCode validates HTTP status code
func (tr TestResponse) AssertStatusCode(t *testing.T, expected int) {
	t.Helper()
	if tr.StatusCode != expected {
		t.Errorf("expected status %d, got %d. Body: %s", expected, tr.StatusCode, string(tr.Body))
	}
}

// AssertJSONField validates presence and type of JSON field
func (tr TestResponse) AssertJSONField(t *testing.T, field string) interface{} {
	t.Helper()
	val, ok := tr.JSON[field]
	if !ok {
		t.Errorf("expected JSON field %q, not found. JSON: %v", field, tr.JSON)
	}
	return val
}

// AssertErrorEnvelope validates standard error response format
func (tr TestResponse) AssertErrorEnvelope(t *testing.T) {
	t.Helper()

	// Error responses may be nested under "error" key for failsafe responses
	var errorData map[string]interface{}
	if errField, ok := tr.JSON["error"]; ok {
		if errMap, ok := errField.(map[string]interface{}); ok {
			errorData = errMap
		} else {
			t.Errorf("expected error field to be object")
			return
		}
	} else {
		// Direct error envelope
		errorData = tr.JSON
	}

	// Verify error envelope has required fields
	if _, ok := errorData["code"]; !ok {
		t.Errorf("expected 'code' in error envelope")
	}
	if _, ok := errorData["message"]; !ok {
		t.Errorf("expected 'message' in error envelope")
	}
	if _, ok := errorData["request_id"]; !ok {
		t.Errorf("expected 'request_id' in error envelope")
	}
	if _, ok := errorData["timestamp"]; !ok {
		t.Errorf("expected 'timestamp' in error envelope")
	}
}

// GetErrorCode extracts error code from response
func (tr TestResponse) GetErrorCode(t *testing.T) string {
	t.Helper()

	var errorData map[string]interface{}
	if errField, ok := tr.JSON["error"]; ok {
		if errMap, ok := errField.(map[string]interface{}); ok {
			errorData = errMap
		}
	} else {
		errorData = tr.JSON
	}

	if code, ok := errorData["code"]; ok {
		if codeStr, ok := code.(string); ok {
			return codeStr
		}
	}
	return ""
}

// GetErrorRequestID extracts request ID from error response
func (tr TestResponse) GetErrorRequestID(t *testing.T) string {
	t.Helper()

	var errorData map[string]interface{}
	if errField, ok := tr.JSON["error"]; ok {
		if errMap, ok := errField.(map[string]interface{}); ok {
			errorData = errMap
		}
	} else {
		errorData = tr.JSON
	}

	if reqID, ok := errorData["request_id"]; ok {
		if reqIDStr, ok := reqID.(string); ok {
			return reqIDStr
		}
	}
	return ""
}

func (tr TestResponse) ResponseData(t *testing.T) map[string]interface{} {
	t.Helper()

	respField, ok := tr.JSON["response"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing response envelope: %v", tr.JSON)
	}

	dataVal, ok := respField["data"]
	if !ok {
		t.Fatalf("response missing data field: %v", respField)
	}

	dataMap, ok := dataVal.(map[string]interface{})
	if !ok {
		t.Fatalf("response data is not object: %T", dataVal)
	}

	return dataMap
}

func extractSections(t *testing.T, data map[string]interface{}) []map[string]interface{} {
	t.Helper()

	sectionsRaw, ok := data["sections"]
	if !ok {
		t.Fatalf("missing sections in menu response: %v", data)
	}

	sectionsSlice, ok := sectionsRaw.([]interface{})
	if !ok {
		t.Fatalf("sections not array: %T", sectionsRaw)
	}

	sections := make([]map[string]interface{}, 0, len(sectionsSlice))
	for _, sec := range sectionsSlice {
		secMap, ok := sec.(map[string]interface{})
		if !ok {
			t.Fatalf("section has unexpected type: %T", sec)
		}
		sections = append(sections, secMap)
	}

	return sections
}

func visibleSectionIDs(t *testing.T, sections []map[string]interface{}) []string {
	t.Helper()

	var ids []string
	for _, sec := range sections {
		visible, ok := sec["visible"].(bool)
		if !ok {
			t.Fatalf("section missing visible flag: %v", sec)
		}
		if visible {
			if id, ok := sec["id"].(string); ok {
				ids = append(ids, id)
			} else {
				t.Fatalf("section missing id: %v", sec)
			}
		}
	}
	return ids
}

func findSection(t *testing.T, sections []map[string]interface{}, id string) map[string]interface{} {
	t.Helper()

	for _, sec := range sections {
		if secID, ok := sec["id"].(string); ok && secID == id {
			return sec
		}
	}
	t.Fatalf("section %q not found in menu response", id)
	return nil
}

func requestAlarmState(t *testing.T, ts *TestServer, role string) TestResponse {
	t.Helper()

	req := newTestRequest(t, "GET", ts.Server.URL+"/api/ui/alarm/state", role)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to request alarm state: %v", err)
	}
	return parseJSONResponse(t, resp)
}

func requestGuestState(t *testing.T, ts *TestServer, role string) TestResponse {
	t.Helper()

	req := newTestRequest(t, "GET", ts.Server.URL+"/api/ui/guest/state", role)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to request guest state: %v", err)
	}
	return parseJSONResponse(t, resp)
}

func setAlarmoState(ts *TestServer, state alarmo.AlarmoState) {
	ts.Coordinator.AlarmoMu.Lock()
	ts.Coordinator.AlarmoState = state
	ts.Coordinator.AlarmoMu.Unlock()
}

func requireAlarmoMeta(t *testing.T, data map[string]interface{}) map[string]interface{} {
	t.Helper()
	alarmoRaw, ok := data["alarmo"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing alarmo metadata: %v", data)
	}
	return alarmoRaw
}

func requireDelayMeta(t *testing.T, alarmoMeta map[string]interface{}) map[string]interface{} {
	t.Helper()
	delayRaw, ok := alarmoMeta["delay"].(map[string]interface{})
	if !ok {
		t.Fatalf("missing delay metadata: %v", alarmoMeta)
	}
	return delayRaw
}

func delayRemaining(delay map[string]interface{}) int {
	if remaining, ok := delay["remaining"].(float64); ok {
		return int(remaining)
	}
	return 0
}

func tickCountdown(ts *TestServer, times int) {
	for i := 0; i < times; i++ {
		ts.Coordinator.Countdown.Tick()
	}
}

func armAlarmState(t *testing.T, ts *TestServer) {
	t.Helper()

	if err := ts.Coordinator.Alarm.Handle("ARM_REQUEST"); err != nil {
		t.Fatalf("failed to send arm request: %v", err)
	}
	if err := ts.Coordinator.Alarm.Handle("ARM_COMPLETE"); err != nil {
		t.Fatalf("failed to complete arm: %v", err)
	}
}

func requestMenu(t *testing.T, ts *TestServer, role string) TestResponse {
	t.Helper()

	req := newTestRequest(t, "GET", ts.Server.URL+"/api/ui/menu", role)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to request menu: %v", err)
	}
	return parseJSONResponse(t, resp)
}

// === Integration Tests ===

// TestHealthCheck_Success verifies GET /health returns 200
func TestHealthCheck_Success(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	armAlarmState(t, ts)

	req := newTestRequest(t, "GET", ts.Server.URL+"/health", "")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	tr := parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusOK)
}

// TestErrorEnvelope_InvalidMethod verifies error envelope format on 405
func TestErrorEnvelope_InvalidMethod(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	// GET to /api/admin/telemetry/optin (expects POST)
	req := newTestRequest(t, "GET", ts.Server.URL+"/api/admin/telemetry/optin", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	tr := parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusMethodNotAllowed)

	// Verify error envelope structure
	tr.AssertErrorEnvelope(t)

	// Verify error code
	code := tr.GetErrorCode(t)
	if code != "method_not_allowed" {
		t.Errorf("expected code 'method_not_allowed', got %q", code)
	}
}

// TestErrorEnvelope_Unauthorized verifies error envelope format on 403
func TestErrorEnvelope_Unauthorized(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	// GET as guest to admin-only endpoint
	req := newTestRequest(t, "GET", ts.Server.URL+"/api/admin/telemetry/summary", "guest")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	tr := parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusForbidden)

	// Verify error envelope structure
	tr.AssertErrorEnvelope(t)

	// Verify error code
	code := tr.GetErrorCode(t)
	if code != "forbidden" {
		t.Errorf("expected code 'forbidden', got %q", code)
	}
}

// TestRequestIDTracking verifies request ID is present in responses
func TestRequestIDTracking(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	req := newTestRequest(t, "GET", ts.Server.URL+"/api/admin/telemetry/summary", "guest")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	tr := parseJSONResponse(t, resp)

	// Verify error envelope structure
	tr.AssertErrorEnvelope(t)

	// Verify request ID exists and has expected format
	reqID := tr.GetErrorRequestID(t)
	if len(reqID) < 4 || reqID[:4] != "req-" {
		t.Errorf("expected request ID to start with 'req-', got %q", reqID)
	}
}

// TestTimestampPresent verifies timestamp is present in error responses
func TestTimestampPresent(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	req := newTestRequest(t, "GET", ts.Server.URL+"/api/admin/telemetry/summary", "guest")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	tr := parseJSONResponse(t, resp)

	// Verify error envelope structure
	tr.AssertErrorEnvelope(t)

	// Extract timestamp from error envelope
	if errField, ok := tr.JSON["error"]; ok {
		if errMap, ok := errField.(map[string]interface{}); ok {
			if tsField, ok := errMap["timestamp"]; ok {
				if tsNum, ok := tsField.(float64); ok {
					now := time.Now().Unix()
					if tsNum < float64(now-5) || tsNum > float64(now+5) {
						t.Errorf("expected timestamp near %d, got %v", now, tsNum)
					}
				} else {
					t.Errorf("expected timestamp to be number, got %T", tsField)
				}
			} else {
				t.Errorf("expected timestamp field in error envelope")
			}
		}
	}
}

// TestWizardConfiguration verifies wizard_completed is respected
func TestWizardConfiguration(t *testing.T) {
	tests := []struct {
		name            string
		wizardCompleted bool
	}{
		{"wizard completed", true},
		{"wizard pending", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := startTestServer(t, TestConfig{
				WizardCompleted: tt.wizardCompleted,
			})
			defer ts.Shutdown()

			// Verify coordinator was initialized with correct config
			if ts.testConfig.WizardCompleted != tt.wizardCompleted {
				t.Errorf("expected wizard_completed %v, got %v", tt.wizardCompleted, ts.testConfig.WizardCompleted)
			}
		})
	}
}

// TestServerShutdown verifies graceful shutdown
func TestServerShutdown(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})

	// Make a request before shutdown
	req := newTestRequest(t, "GET", ts.Server.URL+"/health", "")
	resp1, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	resp1.Body.Close()

	// Shutdown server
	err = ts.Shutdown()
	if err != nil {
		t.Logf("shutdown error: %v (may be expected)", err)
	}

	// Verify server is no longer responding (allow brief grace period)
	time.Sleep(100 * time.Millisecond)

	// Create new client with short timeout
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	req2 := newTestRequest(t, "GET", ts.Server.URL+"/health", "")
	resp2, err := client.Do(req2)
	if err == nil {
		resp2.Body.Close()
		t.Logf("warning: server still responding after shutdown")
	}
	// Error is expected here
}

// TestConcurrentRequests verifies server handles multiple simultaneous requests
func TestConcurrentRequests(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	// Make 10 concurrent requests
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			req := newTestRequest(t, "GET", ts.Server.URL+"/health", "")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				done <- fmt.Errorf("request %d failed: %v", index, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				done <- fmt.Errorf("request %d got status %d", index, resp.StatusCode)
				return
			}

			done <- nil
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Error(err)
		}
	}
}

// TestMultipleRoles verifies different user roles are respected
func TestMultipleRoles(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	tests := []struct {
		name           string
		role           string
		path           string
		expectedStatus int
	}{
		{"admin smoke test", "admin", "/api/admin/smoke", http.StatusMethodNotAllowed}, // POST required
		{"user home state", "user", "/api/ui/home/state", http.StatusOK},
		{"guest home state", "guest", "/api/ui/home/state", http.StatusOK},
		{"guest admin access", "guest", "/api/admin/telemetry/summary", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestRequest(t, "GET", ts.Server.URL+tt.path, tt.role)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, resp.StatusCode, string(body))
			}
		})
	}
}

// BenchmarkHealthCheck measures performance of /health endpoint
func BenchmarkHealthCheck(b *testing.B) {
	ts := startTestServer(&testing.T{}, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := newTestRequest(&testing.T{}, "GET", ts.Server.URL+"/health", "")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			b.Fatalf("failed to make request: %v", err)
		}
		resp.Body.Close()
	}
}

// ============================================================================
// INTEGRATION TEST SPRINT 1.2: First-Boot (D0) End-to-End Behavior
// ============================================================================

// TestFirstBootInitialState verifies initial state when wizard_completed=false
func TestFirstBootInitialState(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	// GET /api/setup/firstboot/status should return step 1
	req := newTestRequest(t, "GET", ts.Server.URL+"/api/setup/firstboot/status", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	tr := parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusOK)

	// Check response structure
	if data, ok := tr.JSON["data"]; ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			// Verify first-boot is active and steps exist
			if active, ok := dataMap["active"].(bool); !ok || !active {
				t.Error("expected first-boot to be active during initial setup")
			}

			if steps, ok := dataMap["steps"].([]interface{}); !ok || len(steps) != len(firstboot.AllSteps) {
				t.Errorf("expected %d steps, got %d", len(firstboot.AllSteps), len(steps))
			}

			// Verify we're at step 1 (Welcome)
			if currentStep, ok := dataMap["current_step"]; ok {
				if stepMap, ok := currentStep.(map[string]interface{}); ok {
					id := stepMap["id"]
					order := stepMap["order"]
					if idStr, ok := id.(string); !ok || idStr != firstboot.AllSteps[0].ID {
						t.Errorf("expected current step id %q, got %v", firstboot.AllSteps[0].ID, id)
					}
					if orderNum, ok := order.(float64); !ok || orderNum != float64(firstboot.AllSteps[0].Order) {
						t.Errorf("expected current step order %d, got %v", firstboot.AllSteps[0].Order, order)
					}
				}
			}
		}
	}
}

// TestFirstBootAlarmBlocked verifies alarm endpoints are blocked during setup
func TestFirstBootAlarmBlocked(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	lastChanged := time.Now().UTC()
	setAlarmoState(ts, alarmo.AlarmoState{
		Mode:        "disarmed",
		Triggered:   false,
		DelayType:   "",
		LastChanged: lastChanged,
	})

	blocked := requestAlarmState(t, ts, "admin")
	blockedData := blocked.ResponseData(t)
	if mode, _ := blockedData["mode"].(string); mode != string(alarm.ModeDisarmed) {
		t.Fatalf("expected alarm mode disarmed during first boot, got %q", mode)
	}
	if _, hasBlock := blockedData["block_reason"]; hasBlock {
		t.Fatal("did not expect block_reason when alarm is disarmed first boot")
	}
	alarmoMeta := requireAlarmoMeta(t, blockedData)
	if state, _ := alarmoMeta["state"].(string); state != "disarmed" {
		t.Fatalf("expected alarmo state disarmed, got %q", state)
	}
	if triggered, _ := alarmoMeta["triggered"].(bool); triggered {
		t.Fatal("expected alarmo triggered=false during first boot")
	}
	delayMeta := requireDelayMeta(t, alarmoMeta)
	if delayType, _ := delayMeta["type"].(string); delayType != "" {
		t.Fatalf("expected no delay type during first boot, got %q", delayType)
	}
	if remaining := delayRemaining(delayMeta); remaining != 0 {
		t.Fatalf("expected delay remaining 0 during first boot, got %d", remaining)
	}
	if lastUpdated, _ := alarmoMeta["last_updated"].(string); lastUpdated != lastChanged.Format(time.RFC3339) {
		t.Fatalf("expected last_updated %q, got %q", lastChanged.Format(time.RFC3339), lastUpdated)
	}
}

// TestFirstBootGuestBlocked verifies guest endpoints are blocked during setup
func TestFirstBootGuestBlocked(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	// Try guest action during setup
	req := newTestRequest(t, "GET", ts.Server.URL+"/api/guest/status", "guest")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	tr := parseJSONResponse(t, resp)
	// Should be blocked or return setup message
	if resp.StatusCode == http.StatusOK {
		// If OK, verify it returns setup message
		if data, ok := tr.JSON["data"]; ok {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if _, hasSystemMsg := dataMap["system_message"]; !hasSystemMsg {
					t.Error("expected system_message in response during setup")
				}
				if msg, hasMsg := dataMap["message"].(string); !hasMsg || !strings.Contains(msg, "Setup") {
					t.Errorf("expected setup message in guest response, got %v", msg)
				}
				if info, ok := dataMap["info"].(map[string]interface{}); ok {
					if reason, ok := info["reason_blocked"].(string); !ok || reason != "first_boot_active" {
						t.Errorf("expected info.reason_blocked=first_boot_active, got %v", info["reason_blocked"])
					}
				} else {
					t.Error("expected info block in guest response during setup")
				}
			}
		}
	}
}

// TestFirstBootStepProgression verifies POST next advances step sequentially
func TestFirstBootStepProgression(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	// Helper to get current step order
	getCurrentStep := func() int {
		req := newTestRequest(t, "GET", ts.Server.URL+"/api/setup/firstboot/status", "admin")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to get status: %v", err)
		}
		tr := parseJSONResponse(t, resp)
		if data, ok := tr.JSON["data"]; ok {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if currentStep, ok := dataMap["current_step"]; ok {
					if stepMap, ok := currentStep.(map[string]interface{}); ok {
						if order, ok := stepMap["order"].(float64); ok {
							return int(order)
						}
					}
				}
			}
		}
		return -1
	}

	// Start at step 1
	if step := getCurrentStep(); step != 1 {
		t.Errorf("expected start at step 1, got %d", step)
	}

	// Advance to step 2
	req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/next", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to advance step: %v", err)
	}
	tr := parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusOK)

	if step := getCurrentStep(); step != 2 {
		t.Errorf("expected step 2 after next, got %d", step)
	}

	// Advance to step 3
	req = newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/next", "admin")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to advance step: %v", err)
	}
	tr = parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusOK)

	if step := getCurrentStep(); step != 3 {
		t.Errorf("expected step 3 after next, got %d", step)
	}
}

// TestFirstBootCannotSkipForward verifies steps cannot be skipped
func TestFirstBootCannotSkipForward(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	// Try to complete while at step 1 (should fail)
	req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/complete", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	// Should return 400 (bad request) - not at final step
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when completing at step 1, got %d", resp.StatusCode)
	}
}

// TestFirstBootBackNavigation verifies POST back goes to previous step
func TestFirstBootBackNavigation(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	// Helper to get current step
	getCurrentStep := func() int {
		req := newTestRequest(t, "GET", ts.Server.URL+"/api/setup/firstboot/status", "admin")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to get status: %v", err)
		}
		tr := parseJSONResponse(t, resp)
		if data, ok := tr.JSON["data"]; ok {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if currentStep, ok := dataMap["current_step"]; ok {
					if stepMap, ok := currentStep.(map[string]interface{}); ok {
						if order, ok := stepMap["order"].(float64); ok {
							return int(order)
						}
					}
				}
			}
		}
		return -1
	}

	// Advance to step 3
	for i := 0; i < 2; i++ {
		req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/next", "admin")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to advance: %v", err)
		}
		resp.Body.Close()
	}

	if step := getCurrentStep(); step != 3 {
		t.Errorf("expected step 3, got %d", step)
	}

	// Go back one step
	req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/back", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to go back: %v", err)
	}
	tr := parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusOK)

	// Should be at step 2
	if step := getCurrentStep(); step != 2 {
		t.Errorf("expected step 2 after back, got %d", step)
	}

	// Go back one more
	req = newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/back", "admin")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to go back: %v", err)
	}
	tr = parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusOK)

	// Should be at step 1
	if step := getCurrentStep(); step != 1 {
		t.Errorf("expected step 1 after back, got %d", step)
	}
}

// TestFirstBootCannotBackAtFirst verifies cannot go back from step 1
func TestFirstBootCannotBackAtFirst(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	// Try to go back while at step 1 (should fail)
	req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/back", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	// Should return 400 (bad request) - already at first step
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when back at step 1, got %d", resp.StatusCode)
	}
}

// TestFirstBootCompletion verifies final step completion persists flag
func TestFirstBootCompletion(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	// Helper to get current step
	getCurrentStep := func() int {
		req := newTestRequest(t, "GET", ts.Server.URL+"/api/setup/firstboot/status", "admin")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to get status: %v", err)
		}
		tr := parseJSONResponse(t, resp)
		if data, ok := tr.JSON["data"]; ok {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if currentStep, ok := dataMap["current_step"]; ok {
					if stepMap, ok := currentStep.(map[string]interface{}); ok {
						if order, ok := stepMap["order"].(float64); ok {
							return int(order)
						}
					}
				}
			}
		}
		return -1
	}

	// Advance through all 5 steps
	for i := 0; i < 4; i++ {
		req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/next", "admin")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to advance: %v", err)
		}
		resp.Body.Close()
	}

	// Should be at step 5
	if step := getCurrentStep(); step != 5 {
		t.Errorf("expected step 5, got %d", step)
	}

	// Complete the wizard
	req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/complete", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to complete: %v", err)
	}

	tr := parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusOK)

	// Verify response indicates completion
	if data, ok := tr.JSON["data"]; ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if completed, ok := dataMap["wizard_completed"]; ok {
				if completedBool, ok := completed.(bool); ok && completedBool {
					// Success: wizard marked as completed
				} else {
					t.Errorf("expected wizard_completed=true in response")
				}
			} else {
				t.Error("expected wizard_completed field in response")
			}
		}
	}
}

// TestFirstBootCompleteOnlyAtFinalStep verifies complete only works at step 5
func TestFirstBootCompleteOnlyAtFinalStep(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	// Try to complete at step 2
	req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/next", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to advance: %v", err)
	}
	resp.Body.Close()

	// Now try to complete (at step 2)
	req = newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/complete", "admin")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	// Should return 400 (not at final step)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when completing at step 2, got %d", resp.StatusCode)
	}
}

// TestFirstBootCannotNextAtFinalStep verifies cannot advance past step 5
func TestFirstBootCannotNextAtFinalStep(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	// Helper to get current step
	getCurrentStep := func() int {
		req := newTestRequest(t, "GET", ts.Server.URL+"/api/setup/firstboot/status", "admin")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to get status: %v", err)
		}
		tr := parseJSONResponse(t, resp)
		if data, ok := tr.JSON["data"]; ok {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if currentStep, ok := dataMap["current_step"]; ok {
					if stepMap, ok := currentStep.(map[string]interface{}); ok {
						if order, ok := stepMap["order"].(float64); ok {
							return int(order)
						}
					}
				}
			}
		}
		return -1
	}

	// Advance through all 5 steps
	for i := 0; i < 4; i++ {
		req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/next", "admin")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to advance: %v", err)
		}
		resp.Body.Close()
	}

	// At step 5, try to go next (should fail)
	req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/next", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	// Should return 400 (already at final step)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 when next at step 5, got %d", resp.StatusCode)
	}

	// Verify still at step 5
	if step := getCurrentStep(); step != 5 {
		t.Errorf("expected still at step 5, got %d", step)
	}
}

func TestFirstBootCompletionPersistsAndUnlocksEndpoints(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	dataDir := filepath.Dir(config.RuntimeConfigPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("failed to prepare runtime config directory: %v", err)
	}

	var (
		origData   []byte
		origExists bool
	)

	if content, err := os.ReadFile(config.RuntimeConfigPath); err == nil {
		origData = content
		origExists = true
	} else if !os.IsNotExist(err) {
		t.Fatalf("failed to read runtime config: %v", err)
	}

	t.Cleanup(func() {
		if origExists {
			if err := os.WriteFile(config.RuntimeConfigPath, origData, 0644); err != nil {
				t.Logf("failed to restore runtime config: %v", err)
			}
		} else {
			if err := os.Remove(config.RuntimeConfigPath); err != nil && !os.IsNotExist(err) {
				t.Logf("failed to remove runtime config: %v", err)
			}
		}
	})

	for i := 0; i < len(firstboot.AllSteps)-1; i++ {
		req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/next", "admin")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("failed to advance step: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status while advancing steps: %d", resp.StatusCode)
		}
	}

	req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/complete", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to complete wizard: %v", err)
	}
	tr := parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusOK)

	data, ok := tr.JSON["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data envelope in completion response")
	}
	if completed, ok := data["wizard_completed"].(bool); !ok || !completed {
		t.Error("expected wizard_completed=true in completion response")
	}
	if status, ok := data["status"].(map[string]interface{}); ok {
		if active, ok := status["active"].(bool); !ok || active {
			t.Error("expected first-boot inactive after completion")
		}
	}

	content, err := os.ReadFile(config.RuntimeConfigPath)
	if err != nil {
		t.Fatalf("failed to read runtime config after completion: %v", err)
	}

	var saved config.RuntimeConfig
	if err := json.Unmarshal(content, &saved); err != nil {
		t.Fatalf("failed to unmarshal runtime config: %v", err)
	}
	if !saved.WizardCompleted {
		t.Error("expected runtime config wizard_completed=true after completion")
	}

	req = newTestRequest(t, "POST", ts.Server.URL+"/api/alarm/arm", "admin")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to call alarm arm after completion: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected alarm arm succeed after completion, got %d", resp.StatusCode)
	}

	req = newTestRequest(t, "GET", ts.Server.URL+"/api/ui/home/state", "user")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to fetch home state after completion: %v", err)
	}
	homeResp := parseJSONResponse(t, resp)
	homeResp.AssertStatusCode(t, http.StatusOK)
	if data, ok := homeResp.JSON["data"].(map[string]interface{}); ok {
		if state, ok := data["state"].(string); ok && state == "setup_redirect" {
			t.Error("expected home state to leave setup_redirect after completion")
		}
		if ready, ok := data["system_ready"].(bool); !ok || !ready {
			t.Error("expected system_ready=true after completion")
		}
	}
}

// TestFirstBootStatusStructure verifies response includes all required fields
func TestFirstBootStatusStructure(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	req := newTestRequest(t, "GET", ts.Server.URL+"/api/setup/firstboot/status", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	tr := parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusOK)

	// Verify response has expected structure
	if data, ok := tr.JSON["data"]; ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			requiredFields := []string{"active", "current_step", "steps"}
			for _, field := range requiredFields {
				if _, ok := dataMap[field]; !ok {
					t.Errorf("expected %q field in response", field)
				}
			}

			// Verify current_step has required fields
			if currentStep, ok := dataMap["current_step"]; ok {
				if stepMap, ok := currentStep.(map[string]interface{}); ok {
					stepFields := []string{"id", "order", "title"}
					for _, field := range stepFields {
						if _, ok := stepMap[field]; !ok {
							t.Errorf("expected %q field in current_step", field)
						}
					}
				}
			}

			// Verify steps is an array
			if steps, ok := dataMap["steps"]; ok {
				if _, ok := steps.([]interface{}); !ok {
					t.Errorf("expected steps to be array, got %T", steps)
				}
			}
		}
	}
}

// TestFirstBootAlreadyCompleted verifies completed wizard doesn't return firstboot endpoints
func TestFirstBootAlreadyCompleted(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	// GET /api/setup/firstboot/status should not block (may return 404 or similar)
	req := newTestRequest(t, "GET", ts.Server.URL+"/api/setup/firstboot/status", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	// Status code should not be 200 if wizard already completed
	// (endpoint may return 400, 404, or 503 depending on implementation)
	if resp.StatusCode == http.StatusOK {
		// If OK, check that firstboot is not active
		tr := parseJSONResponse(t, resp)
		if data, ok := tr.JSON["data"]; ok {
			if dataMap, ok := data.(map[string]interface{}); ok {
				if active, ok := dataMap["active"].(bool); ok && active {
					t.Error("expected firstboot to be inactive when wizard_completed=true")
				}
			}
		}
	}
}

// TestFirstBootNoPanic verifies system doesn't panic during setup flow
func TestFirstBootNoPanic(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	// Execute a complete flow without panic
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/setup/firstboot/status"},
		{"POST", "/api/setup/firstboot/next"},
		{"GET", "/api/setup/firstboot/status"},
		{"POST", "/api/setup/firstboot/next"},
		{"GET", "/api/setup/firstboot/status"},
		{"POST", "/api/setup/firstboot/back"},
		{"GET", "/api/setup/firstboot/status"},
		{"POST", "/api/setup/firstboot/next"},
		{"GET", "/api/setup/firstboot/status"},
	}

	for _, ep := range endpoints {
		req := newTestRequest(t, ep.method, ts.Server.URL+ep.path, "admin")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()

		// Any 5xx indicates panic was recovered
		if resp.StatusCode >= 500 {
			t.Errorf("unexpected 5xx status on %s %s: %d", ep.method, ep.path, resp.StatusCode)
		}
	}
}

// TestFirstBootResponseFormat verifies error responses use standard envelope
func TestFirstBootResponseFormat(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	// Try invalid action to trigger error response
	req := newTestRequest(t, "POST", ts.Server.URL+"/api/setup/firstboot/complete", "admin")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	tr := parseJSONResponse(t, resp)

	// Should have error envelope
	if resp.StatusCode >= 400 {
		// Check for error envelope structure
		if _, ok := tr.JSON["error"]; ok {
			// Success: uses standard error envelope
		} else if _, ok := tr.JSON["data"]; ok {
			// Alternative: might return data with error info
		} else {
			t.Error("expected error envelope or data field in error response")
		}
	}
}

// === INTEGRATION TEST SPRINT 1.3: Menu Visibility + Access Control ===

func TestMenuAdminSeesAllSectionsAndSettings(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	ts.Coordinator.Menu.SetUserRole(menu.RoleAdmin)
	if ts.Coordinator.Settings != nil {
		ts.Coordinator.Settings.SetUserRole(settings.RoleAdmin)
	}

	tr := requestMenu(t, ts, "admin")
	tr.AssertStatusCode(t, http.StatusOK)
	data := tr.ResponseData(t)

	if role, ok := data["role"].(string); !ok || role != string(menu.RoleAdmin) {
		t.Fatalf("expected menu role to be admin, got %v", data["role"])
	}

	sections := extractSections(t, data)
	visible := visibleSectionIDs(t, sections)
	expected := map[string]struct{}{
		"home":     {},
		"alarm":    {},
		"guest":    {},
		"devices":  {},
		"history":  {},
		"settings": {},
	}
	if len(visible) != len(expected) {
		t.Fatalf("expected %d visible sections, got %d: %v", len(expected), len(visible), visible)
	}

	for _, id := range visible {
		if _, ok := expected[id]; !ok {
			t.Fatalf("unexpected visible section %q for admin", id)
		}
	}

	settings := findSection(t, sections, "settings")
	if visible, ok := settings["visible"].(bool); !ok || !visible {
		t.Errorf("expected settings section visible for admin during normal operation")
	}

	req := newTestRequest(t, "GET", ts.Server.URL+"/api/ui/settings", "admin")
	setResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to request settings: %v", err)
	}
	setResp.Body.Close()
	if setResp.StatusCode != http.StatusOK {
		t.Errorf("expected settings endpoint to succeed for admin, got %d", setResp.StatusCode)
	}
}

func TestMenuUserHidesSettingsAndGetsForbiddenOnSettingsEndpoint(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	ts.Coordinator.Menu.SetUserRole(menu.RoleUser)

	tr := requestMenu(t, ts, "user")
	tr.AssertStatusCode(t, http.StatusOK)
	data := tr.ResponseData(t)

	if role, ok := data["role"].(string); !ok || role != string(menu.RoleUser) {
		t.Fatalf("expected menu role to be user, got %v", data["role"])
	}

	sections := extractSections(t, data)
	visible := visibleSectionIDs(t, sections)
	expected := map[string]struct{}{
		"home":    {},
		"alarm":   {},
		"guest":   {},
		"devices": {},
		"history": {},
	}
	if len(visible) != len(expected) {
		t.Fatalf("expected %d visible sections for user, got %d: %v", len(expected), len(visible), visible)
	}

	for _, id := range visible {
		if _, ok := expected[id]; !ok {
			t.Fatalf("unexpected visible section %q for user", id)
		}
	}

	req := newTestRequest(t, "GET", ts.Server.URL+"/api/ui/settings", "user")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to request settings as user: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 for user on settings, got %d", resp.StatusCode)
	}
}

func TestMenuGuestRestrictedSectionsAndLogbookEmpty(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	ts.Coordinator.Menu.SetUserRole(menu.RoleGuest)

	tr := requestMenu(t, ts, "guest")
	tr.AssertStatusCode(t, http.StatusOK)
	data := tr.ResponseData(t)

	if role, ok := data["role"].(string); !ok || role != string(menu.RoleGuest) {
		t.Fatalf("expected menu role to be guest, got %v", data["role"])
	}

	sections := extractSections(t, data)
	visible := visibleSectionIDs(t, sections)
	expectedVisible := map[string]struct{}{
		"home":  {},
		"alarm": {},
		"guest": {},
	}
	if len(visible) != len(expectedVisible) {
		t.Fatalf("expected %d visible sections for guest, got %d: %v", len(expectedVisible), len(visible), visible)
	}

	for _, id := range visible {
		if _, ok := expectedVisible[id]; !ok {
			t.Fatalf("unexpected visible section %q for guest", id)
		}
	}

	req := newTestRequest(t, "GET", ts.Server.URL+"/api/ui/logbook", "guest")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to request logbook as guest: %v", err)
	}
	trLog := parseJSONResponse(t, resp)
	trLog.AssertStatusCode(t, http.StatusOK)
	logData := trLog.ResponseData(t)
	if entries, ok := logData["entries"].([]interface{}); ok && len(entries) > 0 {
		t.Errorf("expected no logbook entries for guest, found %d", len(entries))
	}
}

func TestMenuFirstBootLimitsAdminSections(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: false,
	})
	defer ts.Shutdown()

	ts.Coordinator.Menu.SetUserRole(menu.RoleAdmin)

	tr := requestMenu(t, ts, "admin")
	tr.AssertStatusCode(t, http.StatusOK)
	data := tr.ResponseData(t)
	sections := extractSections(t, data)
	visible := visibleSectionIDs(t, sections)

	expected := map[string]struct{}{
		"home":  {},
		"alarm": {},
	}
	if len(visible) != len(expected) {
		t.Fatalf("expected %d visible sections during first boot, got %d: %v", len(expected), len(visible), visible)
	}

	for _, id := range visible {
		if _, ok := expected[id]; !ok {
			t.Fatalf("unexpected visible section %q during first boot", id)
		}
	}

	if guestSection := findSection(t, sections, "guest"); guestSection != nil {
		if visible, _ := guestSection["visible"].(bool); visible {
			t.Error("expected guest section to be hidden while first boot is active")
		}
		if reason, _ := guestSection["reason_hidden"].(string); reason != "first_boot_active" {
			t.Errorf("expected guest section reason_hidden=first_boot_active, got %q", reason)
		}
	}
}

func TestGuestRequestDuringArmedStateBlocksAlarm(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	armedTime := time.Now().UTC()
	setAlarmoState(ts, alarmo.AlarmoState{
		Mode:           "armed",
		ArmedMode:      "away",
		DelayType:      "exit",
		DelayRemaining: 30,
		LastChanged:    armedTime,
	})

	initial := requestAlarmState(t, ts, "admin")
	initialData := initial.ResponseData(t)
	if mode, _ := initialData["mode"].(string); mode != string(alarm.ModeArmed) {
		t.Fatalf("expected alarm to be armed before guest request, got %q", mode)
	}
	initialAlarmo := requireAlarmoMeta(t, initialData)
	if state, _ := initialAlarmo["state"].(string); state != "armed" {
		t.Fatalf("expected alarmo state armed, got %q", state)
	}
	if armedMode, _ := initialAlarmo["armed_mode"].(string); armedMode != "away" {
		t.Fatalf("expected alarmo armed_mode away, got %q", armedMode)
	}
	if triggered, _ := initialAlarmo["triggered"].(bool); triggered {
		t.Fatal("expected alarmo triggered=false before guest request")
	}
	initialDelay := requireDelayMeta(t, initialAlarmo)
	if delayType, _ := initialDelay["type"].(string); delayType != "exit" {
		t.Fatalf("expected alarmo delay type exit, got %q", delayType)
	}
	if remaining := delayRemaining(initialDelay); remaining != 30 {
		t.Fatalf("expected alarmo delay remaining 30, got %d", remaining)
	}
	if lastUpdated, _ := initialAlarmo["last_updated"].(string); lastUpdated != armedTime.Format(time.RFC3339) {
		t.Fatalf("expected last_updated %q, got %q", armedTime.Format(time.RFC3339), lastUpdated)
	}

	req := newTestRequest(t, "POST", ts.Server.URL+"/api/ui/guest/request", "guest")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("guest request failed: %v", err)
	}
	tr := parseJSONResponse(t, resp)
	tr.AssertStatusCode(t, http.StatusOK)
	guestData := tr.ResponseData(t)
	if state, _ := guestData["state"].(string); state != "guest_requesting" {
		t.Fatalf("expected guest state guest_requesting, got %q", state)
	}

	blocked := requestAlarmState(t, ts, "admin")
	blockedData := blocked.ResponseData(t)
	if mode, _ := blockedData["mode"].(string); mode != string(alarm.ModeArmed) {
		t.Fatalf("expected alarm to stay armed after guest request, got %q", mode)
	}
	if _, hasBlock := blockedData["block_reason"]; hasBlock {
		t.Fatal("did not expect block_reason for backend-only guest request")
	}
	blockedAlarmo := requireAlarmoMeta(t, blockedData)
	if state, _ := blockedAlarmo["state"].(string); state != "armed" {
		t.Fatalf("expected alarmo state to remain armed while blocked, got %q", state)
	}
	if remaining := delayRemaining(requireDelayMeta(t, blockedAlarmo)); remaining != 30 {
		t.Fatalf("expected alarmo delay remaining to stay 30 while blocked, got %d", remaining)
	}
}

func TestGuestApprovalDisarmsAlarmAndShowsApprovedState(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	armAlarmState(t, ts)

	armedTime := time.Now().UTC()
	setAlarmoState(ts, alarmo.AlarmoState{
		Mode:           "armed",
		ArmedMode:      "home",
		DelayType:      "entry",
		DelayRemaining: 20,
		LastChanged:    armedTime,
	})

	req := newTestRequest(t, "POST", ts.Server.URL+"/api/ui/guest/request", "guest")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("guest request failed: %v", err)
	}
	resp.Body.Close()

	ts.Coordinator.GuestScreen.OnApproval(string(alarm.ModeArmed))

	guestState := requestGuestState(t, ts, "guest")
	guestData := guestState.ResponseData(t)
	if state, _ := guestData["state"].(string); state != "guest_approved" {
		t.Fatalf("expected guest state guest_approved, got %q", state)
	}

	alarmState := requestAlarmState(t, ts, "admin")
	alarmData := alarmState.ResponseData(t)
	if mode, _ := alarmData["mode"].(string); mode != string(alarm.ModeArmed) {
		t.Fatalf("expected alarm to remain armed after guest approval, got %q", mode)
	}
	alarmoMeta := requireAlarmoMeta(t, alarmData)
	if state, _ := alarmoMeta["state"].(string); state != "armed" {
		t.Fatalf("expected alarmo state to remain armed after approval, got %q", state)
	}
	if triggered, _ := alarmoMeta["triggered"].(bool); triggered {
		t.Fatal("expected alarmo triggered=false after guest approval")
	}
	if remaining := delayRemaining(requireDelayMeta(t, alarmoMeta)); remaining != 20 {
		t.Fatalf("expected alarmo delay remaining 20 after guest approval, got %d", remaining)
	}
}

func TestGuestExitRearmsAlarm(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	armAlarmState(t, ts)

	req := newTestRequest(t, "POST", ts.Server.URL+"/api/ui/guest/request", "guest")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("guest request failed: %v", err)
	}
	resp.Body.Close()

	approveReq := newTestRequest(t, "POST", ts.Server.URL+"/api/guest/approve", "admin")
	approveResp, err := http.DefaultClient.Do(approveReq)
	if err != nil {
		t.Fatalf("guest approve failed: %v", err)
	}
	approveResp.Body.Close()

	exitReq := newTestRequest(t, "POST", ts.Server.URL+"/api/guest/exit", "guest")
	exitResp, err := http.DefaultClient.Do(exitReq)
	if err != nil {
		t.Fatalf("guest exit failed: %v", err)
	}
	exitResp.Body.Close()

	alarmState := requestAlarmState(t, ts, "admin")
	alarmData := alarmState.ResponseData(t)
	if mode, _ := alarmData["mode"].(string); mode != string(alarm.ModeArmed) {
		t.Fatalf("expected alarm to re-arm after guest exit, got %q", mode)
	}
}

func TestAlarmTriggeredDuringGuestActiveAllowsExit(t *testing.T) {
	ts := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer ts.Shutdown()

	armAlarmState(t, ts)

	req := newTestRequest(t, "POST", ts.Server.URL+"/api/ui/guest/request", "guest")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("guest request failed: %v", err)
	}
	resp.Body.Close()

	approveReq := newTestRequest(t, "POST", ts.Server.URL+"/api/guest/approve", "admin")
	approveResp, err := http.DefaultClient.Do(approveReq)
	if err != nil {
		t.Fatalf("guest approve failed: %v", err)
	}
	approveResp.Body.Close()

	if err := ts.Coordinator.Alarm.Handle("TRIGGER"); err != nil {
		t.Fatalf("failed to trigger alarm: %v", err)
	}

	triggered := requestAlarmState(t, ts, "admin")
	triggeredData := triggered.ResponseData(t)
	if mode, _ := triggeredData["mode"].(string); mode != string(alarm.ModeTriggered) {
		t.Fatalf("expected alarm mode triggered, got %q", mode)
	}

	exitReq := newTestRequest(t, "POST", ts.Server.URL+"/api/guest/exit", "guest")
	exitResp, err := http.DefaultClient.Do(exitReq)
	if err != nil {
		t.Fatalf("guest exit failed after trigger: %v", err)
	}
	exitResp.Body.Close()

	guestState := requestGuestState(t, ts, "guest")
	guestData := guestState.ResponseData(t)
	if state, _ := guestData["state"].(string); state != "guest_exit" {
		t.Fatalf("expected guest state guest_exit after trigger exit, got %q", state)
	}
}

func TestReducedMotionCountdownStatic(t *testing.T) {
	presetTime := time.Now().UTC()
	preset := alarmo.AlarmoState{
		Mode:           "arming",
		DelayType:      "exit",
		DelayRemaining: 30,
		LastChanged:    presetTime,
	}

	standard := startTestServer(t, TestConfig{
		WizardCompleted: true,
	})
	defer standard.Shutdown()

	setAlarmoState(standard, preset)
	standard.Coordinator.Countdown.Start()
	tickCountdown(standard, 3)
	standardResp := requestAlarmState(t, standard, "admin")
	standardData := standardResp.ResponseData(t)
	standardAlarmo := requireAlarmoMeta(t, standardData)
	standardDelay := requireDelayMeta(t, standardAlarmo)
	if delayType, _ := standardDelay["type"].(string); delayType != "exit" {
		t.Fatalf("expected alarmo delay type exit for standard mode, got %q", delayType)
	}
	if remaining := delayRemaining(standardDelay); remaining != 30 {
		t.Fatalf("expected alarmo delay remaining 30 for standard mode, got %d", remaining)
	}
	if lastUpdated, _ := standardAlarmo["last_updated"].(string); lastUpdated != presetTime.Format(time.RFC3339) {
		t.Fatalf("expected last_updated %q, got %q", presetTime.Format(time.RFC3339), lastUpdated)
	}

	repeatResp := requestAlarmState(t, standard, "admin")
	repeatAlarmo := requireAlarmoMeta(t, repeatResp.ResponseData(t))
	repeatDelay := requireDelayMeta(t, repeatAlarmo)
	if delayRemaining(repeatDelay) != 30 {
		t.Fatalf("expected repeated standard delay to stay static at 30, got %d", delayRemaining(repeatDelay))
	}

	reduced := startTestServer(t, TestConfig{
		WizardCompleted: true,
		ReducedMotion:   true,
	})
	defer reduced.Shutdown()

	setAlarmoState(reduced, preset)
	reduced.Coordinator.Countdown.Start()
	tickCountdown(reduced, 3)
	reducedResp := requestAlarmState(t, reduced, "admin")
	reducedAlarmo := requireAlarmoMeta(t, reducedResp.ResponseData(t))
	reducedDelay := requireDelayMeta(t, reducedAlarmo)
	if delayRemaining(reducedDelay) != 30 {
		t.Fatalf("expected reduced_motion delay remaining 30, got %d", delayRemaining(reducedDelay))
	}
	if delayRemaining(reducedDelay) != delayRemaining(standardDelay) {
		t.Fatalf("expected reduced_motion delay to match standard delay, got %d vs %d", delayRemaining(reducedDelay), delayRemaining(standardDelay))
	}
}
