/**
 * SmartDisplay Menu Controller
 * Manages menu state, routing, and role-based visibility
 */

(function() {
    'use strict';

    // ========================================================================
    // Menu Controller
    // ========================================================================
    var MenuController = {
        menuData: null,
        isLoading: false,
        error: null,
        lastUpdateTime: null,

        // ====================================================================
        // Fetch Menu Data
        // ====================================================================

        /**
         * Load menu structure from backend
         * @returns {Promise<object>} - Menu data
         */
        fetchMenu: function() {
            var self = this;

            console.log('[Menu] Fetching menu...');

            return window.SmartDisplay.api.client.get('/ui/menu', {
                headers: {
                    'X-User-Role': 'admin'
                }
            })
                .then(function(response) {
                    console.log('[Menu] Menu loaded:', response);
                    self.menuData = response;
                    self.error = null;
                    self.lastUpdateTime = Date.now();

                    return response;
                })
                .catch(function(err) {
                    console.error('[Menu] Failed to fetch menu:', err);
                    self.error = err;
                    throw err;
                });
        },

        /**
         * Get currently visible sections
         * @returns {array} - Array of visible section objects
         */
        getVisibleSections: function() {
            if (!this.menuData || !this.menuData.sections) {
                return [];
            }

            // Filter to only visible sections
            return this.menuData.sections.filter(function(section) {
                return section.visible !== false;
            });
        },

        /**
         * Get menu item by view ID
         * @param {string} viewId - View identifier
         * @returns {object|null} - Menu item or null
         */
        getMenuItemByView: function(viewId) {
            var sections = this.getVisibleSections();

            for (var i = 0; i < sections.length; i++) {
                var section = sections[i];
                if (section.items && Array.isArray(section.items)) {
                    for (var j = 0; j < section.items.length; j++) {
                        var item = section.items[j];
                        if (item.view === viewId) {
                            return item;
                        }
                    }
                }
            }

            return null;
        },

        /**
         * Check if a menu item is enabled
         * @param {object} item - Menu item
         * @returns {boolean}
         */
        isItemEnabled: function(item) {
            return item.enabled !== false;
        },

        // ====================================================================
        // Polling Setup
        // ====================================================================

        /**
         * Setup polling provider for store
         */
        setupPolling: function() {
            var self = this;

            console.log('[Menu] Setting up polling provider');

            // Register polling provider
            window.SmartDisplay.store.registerPollingProvider(function() {
                return self.fetchMenu()
                    .catch(function(err) {
                        console.error('[Menu] Polling error:', err);
                        return null;
                    });
            });
        },

        // ====================================================================
        // State Subscription
        // ====================================================================

        /**
         * Subscribe to state changes that affect menu
         */
        subscribeToStateChanges: function() {
            var self = this;

            console.log('[Menu] Subscribing to state changes');

            // Subscribe to store changes
            if (window.SmartDisplay.store) {
                window.SmartDisplay.store.subscribe(function(updates) {
                    // Check if updates that affect menu
                    if (updates.firstBoot !== undefined ||
                        (updates.guestState && updates.guestState.state) ||
                        (updates.alarmState && updates.alarmState.status)) {
                        
                        console.log('[Menu] State change detected, refreshing menu');
                        self.fetchMenu()
                            .catch(function(err) {
                                console.error('[Menu] Failed to refresh menu:', err);
                            });
                    }
                });
            }
        },

        // ====================================================================
        // Initialization
        // ====================================================================

        /**
         * Initialize controller
         */
        init: function() {
            var self = this;

            console.log('[Menu] Initializing controller');

            // Setup polling
            this.setupPolling();

            // Subscribe to state changes
            this.subscribeToStateChanges();

            // Load initial menu
            return this.fetchMenu()
                .catch(function(err) {
                    console.error('[Menu] Failed to initialize:', err);
                });
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.menuController = MenuController;

    console.log('[SmartDisplay] Menu controller registered');

})();
