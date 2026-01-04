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
                    console.log('[Alarm] Alarm state loaded:', response);
                    self.currentState = response;
                    self.error = null;
                    self.lastUpdateTime = Date.now();
                    
                    // Update store with alarm data
                    if (response.alarmState) {
                        window.SmartDisplay.store.setState({
                            alarmState: response.alarmState
                        });
                    }

                    return response;
                })
                .catch(function(err) {
                    console.error('[Alarm] Failed to fetch state:', err);
                    self.error = err;
                    throw err;
                });
        },

        // ====================================================================
        // Alarm Actions
        // ====================================================================

        /**
         * Request alarm action
         * @param {string} action - Action name (arm-home, arm-away, disarm, etc.)
         * @returns {Promise<object>} - Action response
         */
        performAction: function(action) {
            var self = this;

            if (this.isLoading) {
                console.warn('[Alarm] Already loading, skipping action');
                return Promise.reject({ message: 'Already processing' });
            }

            this.isLoading = true;
            console.log('[Alarm] Performing action:', action);

            return window.SmartDisplay.api.client.post('/ui/alarm/action', {
                action: action
            })
                .then(function(response) {
                    console.log('[Alarm] Action succeeded:', action, response);
                    self.isLoading = false;
                    
                    // Fetch updated state
                    return self.fetchAlarmState();
                })
                .catch(function(err) {
                    console.error('[Alarm] Action failed:', action, err);
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
