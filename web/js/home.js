/**
 * SmartDisplay Home View Controller
 * Manages home/idle screen data, polling, and state
 */

(function() {
    'use strict';

    // ========================================================================
    // Home Controller
    // ========================================================================
    var HomeController = {
        currentState: null,
        isActive: false,
        error: null,
        lastUpdateTime: null,

        // ====================================================================
        // Fetch Home State
        // ====================================================================

        /**
         * Load home state from backend
         * @returns {Promise<object>} - Home state data
         */
        fetchHomeState: function() {
            var self = this;

            console.log('[Home] Fetching home state...');

            return window.SmartDisplay.api.client.get('/ui/home/state')
                .then(function(response) {
                    // API responses are wrapped: { response: { ok, data }, failsafe }
                    var payload = response && response.response && response.response.data
                        ? response.response.data
                        : (response && response.data) ? response.data : response;

                    console.log('[Home] Home state loaded:', payload);
                    self.currentState = payload;
                    self.error = null;
                    self.lastUpdateTime = Date.now();
                    
                    // Update store with mapped home summary (best-effort)
                    if (payload && payload.summary) {
                        var summary = payload.summary;
                        window.SmartDisplay.store.setState({
                            homeState: {
                                aiInsight: summary.ai_insight || '',
                                aiSeverity: summary.countdown_active ? 'Countdown' : '',
                                aiOneLiner: payload.message || summary.ai_insight || '',
                                temperature: summary.temperature !== undefined ? summary.temperature : null
                            }
                        });
                    }
                    
                    return payload;
                })
                .catch(function(err) {
                    console.error('[Home] Failed to fetch state:', err);
                    self.error = err;
                    throw err;
                });
        },

        // ====================================================================
        // Active State
        // ====================================================================

        /**
         * Mark as active (user interaction)
         */
        setActive: function() {
            if (!this.isActive) {
                this.isActive = true;
                console.log('[Home] Marked as active');
                
                // Dispatch event for views to listen to
                var event = new CustomEvent('home-active', {
                    detail: { timestamp: Date.now() }
                });
                document.dispatchEvent(event);
            }
        },

        /**
         * Mark as inactive (idle timeout)
         */
        setInactive: function() {
            if (this.isActive) {
                this.isActive = false;
                console.log('[Home] Marked as inactive');
                
                var event = new CustomEvent('home-inactive', {
                    detail: { timestamp: Date.now() }
                });
                document.dispatchEvent(event);
            }
        },

        // ====================================================================
        // Polling Setup
        // ====================================================================

        /**
         * Setup polling provider for store
         */
        setupPolling: function() {
            var self = this;

            console.log('[Home] Setting up polling provider');

            // Register polling provider
            window.SmartDisplay.store.registerPollingProvider(function() {
                return self.fetchHomeState()
                    .catch(function(err) {
                        console.error('[Home] Polling error:', err);
                        // Return empty object to continue polling
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

            console.log('[Home] Initializing controller');

            // Setup polling
            this.setupPolling();

            // Load initial state
            return this.fetchHomeState()
                .catch(function(err) {
                    console.error('[Home] Failed to initialize:', err);
                    // Continue with error state
                });
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.homeController = HomeController;

    console.log('[SmartDisplay] Home controller registered');

})();
