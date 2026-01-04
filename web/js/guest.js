/**
 * SmartDisplay Guest View Controller
 * Manages guest access state, polling, and actions
 */

(function() {
    'use strict';

    // ========================================================================
    // Guest Controller
    // ========================================================================
    var GuestController = {
        currentState: null,
        isLoading: false,
        error: null,
        lastUpdateTime: null,

        // ====================================================================
        // Fetch Guest State
        // ====================================================================

        /**
         * Load guest state from backend
         * @returns {Promise<object>} - Guest state data
         */
        fetchGuestState: function() {
            var self = this;

            console.log('[Guest] Fetching guest state...');

            return window.SmartDisplay.api.client.get('/ui/guest/state')
                .then(function(response) {
                    console.log('[Guest] Guest state loaded:', response);
                    self.currentState = response;
                    self.error = null;
                    self.lastUpdateTime = Date.now();
                    
                    // Update store with guest data
                    if (response.guestState) {
                        window.SmartDisplay.store.setState({
                            guestState: response.guestState
                        });
                    }

                    return response;
                })
                .catch(function(err) {
                    console.error('[Guest] Failed to fetch state:', err);
                    self.error = err;
                    throw err;
                });
        },

        // ====================================================================
        // Guest Actions
        // ====================================================================

        /**
         * Request guest access
         * @returns {Promise<object>} - Request response
         */
        requestAccess: function() {
            var self = this;

            if (this.isLoading) {
                console.warn('[Guest] Already loading, skipping request');
                return Promise.reject({ message: 'Already processing' });
            }

            this.isLoading = true;
            console.log('[Guest] Requesting access...');

            return window.SmartDisplay.api.client.post('/ui/guest/request', {})
                .then(function(response) {
                    console.log('[Guest] Request submitted:', response);
                    self.isLoading = false;
                    
                    // Fetch updated state
                    return self.fetchGuestState();
                })
                .catch(function(err) {
                    console.error('[Guest] Request failed:', err);
                    self.error = err;
                    self.isLoading = false;
                    throw err;
                });
        },

        /**
         * Exit guest mode
         * @returns {Promise<object>} - Exit response
         */
        exitGuest: function() {
            var self = this;

            if (this.isLoading) {
                console.warn('[Guest] Already loading, skipping exit');
                return Promise.reject({ message: 'Already processing' });
            }

            this.isLoading = true;
            console.log('[Guest] Exiting guest mode...');

            return window.SmartDisplay.api.client.post('/ui/guest/exit', {})
                .then(function(response) {
                    console.log('[Guest] Exit successful:', response);
                    self.isLoading = false;
                    
                    // Fetch updated state
                    return self.fetchGuestState();
                })
                .catch(function(err) {
                    console.error('[Guest] Exit failed:', err);
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

            console.log('[Guest] Setting up polling provider');

            // Register polling provider
            window.SmartDisplay.store.registerPollingProvider(function() {
                return self.fetchGuestState()
                    .catch(function(err) {
                        console.error('[Guest] Polling error:', err);
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

            console.log('[Guest] Initializing controller');

            // Setup polling
            this.setupPolling();

            // Load initial state
            return this.fetchGuestState()
                .catch(function(err) {
                    console.error('[Guest] Failed to initialize:', err);
                });
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.guestController = GuestController;

    console.log('[SmartDisplay] Guest controller registered');

})();
