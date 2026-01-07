/**
 * SmartDisplay Settings Controller
 * FAZ S4: Manages Home Assistant connection state polling and display
 */

(function() {
    'use strict';

    var Settings = {
        // ====================================================================
        // Public API
        // ====================================================================

        /**
         * Initialize settings controller
         */
        init: function() {
            console.log('[Settings] Initializing controller');

            // Setup polling
            this.setupPolling();

            // Load initial HA status
            return this.fetchHAStatus()
                .catch(function(err) {
                    console.error('[Settings] Failed to initialize:', err);
                    return null;
                });
        },

        /**
         * Setup polling for HA connection status
         */
        setupPolling: function() {
            var self = this;

            console.log('[Settings] Setting up HA status polling provider');

            // Register polling provider for HA connection status
            window.SmartDisplay.store.registerPollingProvider(function() {
                return self.fetchHAStatus()
                    .catch(function(err) {
                        console.error('[Settings] HA status polling error:', err);
                        // Return empty object to continue polling
                        return null;
                    });
            });
        },

        /**
         * Fetch current HA connection status from API
         * FAZ S4: Safe endpoint that returns connection state without secrets
         */
        fetchHAStatus: function() {
            var self = this;

            return window.SmartDisplay.api.client.get('/settings/homeassistant/status', {
                headers: {
                    'X-User-Role': 'admin'
                }
            })
            .then(function(envelope) {
                // Handle successful response
                // Response is wrapped: { failsafe: {...}, response: { ok: true, data: {...} } }
                var response = envelope.response || {};
                if (!response.ok) {
                    throw new Error('HA status API error: ' + (response.error || 'unknown'));
                }

                var data = response.data || {};

                // Map API response to store state
                var updates = {
                    haState: {
                        isConnected: data.ha_connected || false,
                        lastTestedAt: data.ha_last_tested_at || null,
                        isConfigured: data.is_configured || false,
                        configuredAt: data.configured_at || null,
                        syncDone: data.initial_sync_done || false,
                        syncAt: data.initial_sync_at || null,
                        meta: data.ha_meta ? {
                            version: data.ha_meta.version || null,
                            timeZone: data.ha_meta.time_zone || null,
                            locationName: data.ha_meta.location_name || null
                        } : null,
                        entityCounts: data.entity_counts ? {
                            lights: data.entity_counts.lights || 0,
                            sensors: data.entity_counts.sensors || 0,
                            switches: data.entity_counts.switches || 0,
                            others: data.entity_counts.others || 0
                        } : null,
                        // FAZ S6: Runtime health state
                        runtimeUnreachable: data.ha_runtime_unreachable || false,
                        lastSeenAt: data.ha_last_seen_at || null
                    }
                };

                window.SmartDisplay.store.setState(updates);

                // FAZ L4: Check advisor hints
                if (window.SmartDisplay.advisor) {
                    var state = window.SmartDisplay.store.getState();
                    window.SmartDisplay.advisor.checkAndShow({
                        role: state.authState.role,
                        alarmState: state.alarmState.state,
                        haIsConnected: state.haState.isConnected,
                        haIsConfigured: state.haState.isConfigured,
                        haSyncDone: state.haState.syncDone,
                        guestIsActive: state.guestState.active,
                        guestApprovedAt: state.guestState.approvalTime,
                        currentView: state.menu.currentView
                    });
                }

                return updates;
            })
            .catch(function(err) {
                // Handle fetch failure gracefully - don't abort Settings init
                console.error('[Settings] HA status fetch failed:', err);
                return {
                    haState: {
                        isConnected: false,
                        lastTestedAt: null,
                        isConfigured: false,
                        configuredAt: null,
                        syncDone: false,
                        syncAt: null,
                        meta: null,
                        entityCounts: null,
                        runtimeUnreachable: false,
                        lastSeenAt: null
                    }
                };
            });
        },

        /**
         * Perform initial HA synchronization
         * FAZ S5: One-time bootstrap sync
         */
        performSync: function() {
            console.log('[Settings] Performing initial HA synchronization');

            return window.SmartDisplay.api.client.post('/settings/homeassistant/sync', null, {
                headers: {
                    'X-User-Role': 'admin'
                }
            })
            .then(function(envelope) {
                var response = envelope.response || envelope;
                if (!response.ok) {
                    throw new Error('HA sync error: ' + (response.error || envelope.error || 'unknown'));
                }

                var syncResult = response.data || envelope.data || {};
                console.log('[Settings] HA sync result:', syncResult.success ? 'success' : 'failed');

                // FAZ L6: Add trace entry
                if (window.SmartDisplay.trace) {
                    window.SmartDisplay.trace.add('Initial HA sync completed');
                }

                // Immediately fetch updated status
                return this.fetchHAStatus();
            }.bind(this));
        },

        /**
         * Test HA connection
         * FAZ S4: Verify HA is reachable with current credentials
         */
        testHAConnection: function() {
            console.log('[Settings] Testing HA connection');

            return window.SmartDisplay.api.client.post('/settings/homeassistant/test', null, {
                headers: {
                    'X-User-Role': 'admin'
                }
            })
            .then(function(envelope) {
                var response = envelope.response || envelope;
                if (!response.ok) {
                    throw new Error('HA test error: ' + (response.error || envelope.error || 'unknown'));
                }

                var testResult = response.data || envelope.data || {};
                console.log('[Settings] HA connection test result:', testResult.stage);

                // FAZ L6: Add trace entry for successful test
                if (window.SmartDisplay.trace) {
                    window.SmartDisplay.trace.add('HA connection verified');
                }

                // Immediately fetch updated status
                return this.fetchHAStatus();
            }.bind(this));
        },

        /**
         * Save HA credentials
         * FAZ S2/S4: Admin-only operation
         */
        saveCredentials: function(serverUrl, token) {
            console.log('[Settings] Saving HA credentials');
            console.log('[Settings] Server URL:', serverUrl);
            console.log('[Settings] Token length:', token.length);

            if (!serverUrl || !token) {
                return Promise.reject(new Error('Server URL and token are required'));
            }

            return window.SmartDisplay.api.client.post('/settings/homeassistant', {
                server_url: serverUrl,
                token: token
            }, {
                headers: {
                    'X-User-Role': 'admin'
                }
            })
            .then(function(envelope) {
                // Handle nested response structure
                var response = envelope.response || envelope;
                if (!response.ok) {
                    throw new Error('HA credentials save error: ' + (response.error || envelope.error || 'unknown'));
                }

                console.log('[Settings] HA credentials saved successfully');

                // FAZ L6: Add trace entry
                if (window.SmartDisplay.trace) {
                    window.SmartDisplay.trace.add('HA credentials saved');
                }

                // Return the response data for frontend form update
                var responseData = response.data || response;
                
                // Automatically test connection after saving credentials
                console.log('[Settings] Auto-testing connection after save...');
                return this.testHAConnection()
                    .then(function() {
                        // Return response data along with test result
                        return responseData;
                    });
            }.bind(this));
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.settings = Settings;

    // Auto-initialize when DOM ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', function() {
            Settings.init().catch(function(err) {
                console.error('[Settings] Initialization failed:', err);
            });
        });
    } else {
        Settings.init().catch(function(err) {
            console.error('[Settings] Initialization failed:', err);
        });
    }
})();
