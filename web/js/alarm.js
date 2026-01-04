/**
 * SmartDisplay Alarm View Controller
 * Manages alarm state, polling, and action dispatch
 */

(function() {
    'use strict';

    // ========================================================================
    // Alarm Controller
    // ========================================================================
    var AlarmController = {
        currentState: null,
        isLoading: false,
        error: null,
        lastUpdateTime: null,
        pollFailureCount: 0,
        connectionWarningShown: false,

        // ====================================================================
        // Fetch Alarm State
        // ====================================================================

        /**
         * Load alarm state from backend
         * @returns {Promise<object>} - Alarm state data
         */
        fetchAlarmState: function() {
            var self = this;

            return window.SmartDisplay.api.client.get('/ui/alarm/state')
                .then(function(response) {
                    var normalized = self._normalizeState(response);
                    console.log('[Alarm] State updated');
                    self.currentState = normalized;
                    self.error = null;
                    self.lastUpdateTime = Date.now();
                    
                    // A5.3: Reset poll failure tracking on success
                    if (self.pollFailureCount > 0) {
                        console.log('[Alarm] Connection recovered');
                    }
                    self.pollFailureCount = 0;
                    self.connectionWarningShown = false;

                    return {
                        alarmState: normalized
                    };
                })
                .catch(function(err) {
                    // A5.3: Track poll failures
                    self.pollFailureCount++;
                    self.error = err;
                    
                    if (self.pollFailureCount === 1) {
                        console.log('[Alarm] Poll failed, retrying...');
                    } else if (!self.connectionWarningShown && self.pollFailureCount >= 3) {
                        console.log('[Alarm] Connection issue detected');
                        self.connectionWarningShown = true;
                    }
                    
                    throw err;
                });
        },

        /**
         * Normalize Alarmo response into store shape
         * @private
         */
        _normalizeState: function(response) {
            return {
                state: (response && response.state) || 'unknown',
                triggered: Boolean(response && response.triggered),
                delay: this._normalizeDelay(response && response.delay),
                lastUpdated: (response && response.last_updated) || new Date().toISOString(),
                isHydrated: true
            };
        },

        _normalizeDelay: function(delay) {
            if (!delay || typeof delay.remaining !== 'number') {
                return null;
            }

            var type = delay.type === 'entry' ? 'entry' : 'exit';
            var remaining = Math.max(0, Math.floor(delay.remaining));

            return {
                type: type,
                remaining: remaining
            };
        },

        // ====================================================================
        // Alarm Actions (A4)
        // ====================================================================

        /**
         * Request alarm action (arm/disarm)
         * @param {string} action - One of: arm_home, arm_away, arm_night, disarm
         * @returns {Promise<object>} - Action response
         */
        requestAction: function(action) {
            var self = this;

            if (this.isLoading) {
                console.warn('[Alarm] Already loading, skipping action');
                return Promise.reject({ message: 'Already processing' });
            }

            var validActions = ['arm_home', 'arm_away', 'arm_night', 'disarm'];
            if (validActions.indexOf(action) === -1) {
                console.error('[Alarm] Invalid action:', action);
                return Promise.reject({ message: 'Invalid action' });
            }

            this.isLoading = true;
            console.log('[Alarm] Action:', action);

            return window.SmartDisplay.api.client.post('/ui/alarm/action', { action: action })
                .then(function(response) {
                    console.log('[Alarm] Action accepted');
                    self.isLoading = false;
                    self.error = null;
                    return response;
                })
                .catch(function(err) {
                    console.log('[Alarm] Action failed:', err.statusCode || 'unknown');
                    self.error = err;
                    self.isLoading = false;
                    throw err;
                });
        },

        // ====================================================================
        // Polling Setup
        // ====================================================================

        /**
         * Setup polling provider for store
         */
        setupPolling: function() {
            var self = this;

            console.log('[Alarm] Setting up polling provider');

            // Register polling provider
            window.SmartDisplay.store.registerPollingProvider(function() {
                return self.fetchAlarmState()
                    .catch(function(err) {
                        console.error('[Alarm] Polling error:', err);
                        return null;
                    });
            });
        },

        // ====================================================================
        // Initialization
        // ====================================================================

        /**
         * Initialize controller
         */
        init: function() {
            var self = this;

            console.log('[Alarm] Initializing controller');

            // Setup polling
            this.setupPolling();

            // Load initial state
            return this.fetchAlarmState()
                .catch(function(err) {
                    console.error('[Alarm] Failed to initialize:', err);
                });
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.alarmController = AlarmController;

    console.log('[SmartDisplay] Alarm controller registered');

})();
