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

        // ====================================================================
        // Fetch Alarm State
        // ====================================================================

        /**
         * Load alarm state from backend
         * @returns {Promise<object>} - Alarm state data
         */
        fetchAlarmState: function() {
            var self = this;

            console.log('[Alarm] Fetching alarm state...');

            return window.SmartDisplay.api.client.get('/ui/alarm/state')
                .then(function(response) {
                    var normalized = self._normalizeState(response);
                    console.log('[Alarm] Alarm state loaded:', normalized);
                    self.currentState = normalized;
                    self.error = null;
                    self.lastUpdateTime = Date.now();

                    return {
                        alarmState: normalized
                    };
                })
                .catch(function(err) {
                    console.error('[Alarm] Failed to fetch state:', err);
                    self.error = err;
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
