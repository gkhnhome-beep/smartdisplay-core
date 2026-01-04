/**
 * SmartDisplay State Store
 * Centralized state management with polling support
 */

(function() {
    'use strict';

    // ========================================================================
    // Store Configuration
    // ========================================================================
    var DEFAULT_POLL_INTERVAL = 5000; // 5 seconds
    var DEFAULT_FAST_POLL_INTERVAL = 1000; // 1 second during critical states
    var CRITICAL_STATES = ['arming', 'pending', 'triggered'];

    // ========================================================================
    // State Store
    // ========================================================================
    var Store = {
        // Global state tree
        state: {
            // First boot flag
            firstBoot: false,

            // Home view state
            homeState: {
                aiInsight: null,
                aiSeverity: null,
                lights: [],
                covers: [],
                plugs: [],
                temperature: null
            },

            // Alarm subsystem state
            alarmState: {
                state: 'unknown',
                triggered: false,
                delay: null,
                lastUpdated: null,
                isHydrated: false
            },

            // Guest access state
            guestState: {
                isGuestActive: false,
                requestPending: false,
                approved: false,
                denyReason: null,
                countdownSeconds: 0
            },

            // Menu/navigation state
            menu: {
                currentView: 'home', // 'home', 'alarm', 'devices', 'settings'
                previousView: null
            }
        },

        // Polling management
        _polling: {
            active: false,
            interval: DEFAULT_POLL_INTERVAL,
            timerHandle: null,
            paused: false,
            providers: [] // Array of polling provider functions
        },

        // State subscribers for change notifications
        _subscribers: [],

        // ====================================================================
        // Public API: State Access
        // ====================================================================

        /**
         * Get entire state tree (shallow copy)
         * @returns {object} - Copy of current state
         */
        getState: function() {
            return JSON.parse(JSON.stringify(this.state));
        },

        /**
         * Get specific state branch
         * @param {string} path - State path (e.g., 'alarmState.status')
         * @returns {*} - State value or undefined
         */
        getStatePath: function(path) {
            var parts = path.split('.');
            var current = this.state;

            for (var i = 0; i < parts.length; i++) {
                if (current && typeof current === 'object' && parts[i] in current) {
                    current = current[parts[i]];
                } else {
                    return undefined;
                }
            }

            return current;
        },

        /**
         * Update state - merges with existing state
         * @param {object} updates - Partial state updates
         * @returns {void}
         */
        setState: function(updates) {
            var changed = false;

            // Deep merge updates into state
            function deepMerge(target, source) {
                var hasChanges = false;

                for (var key in source) {
                    if (source.hasOwnProperty(key)) {
                        if (source[key] === null || source[key] === undefined) {
                            if (target[key] !== source[key]) {
                                target[key] = source[key];
                                hasChanges = true;
                            }
                        } else if (typeof source[key] === 'object' && !Array.isArray(source[key])) {
                            if (!(key in target) || typeof target[key] !== 'object') {
                                target[key] = {};
                            }
                            if (deepMerge(target[key], source[key])) {
                                hasChanges = true;
                            }
                        } else {
                            if (target[key] !== source[key]) {
                                target[key] = source[key];
                                hasChanges = true;
                            }
                        }
                    }
                }

                return hasChanges;
            }

            if (deepMerge(this.state, updates)) {
                changed = true;
            }

            if (changed) {
                // Check if we should adjust polling speed
                this._updatePollingSpeed();

                // Notify subscribers
                this._notifySubscribers(updates);
            }
        },

        /**
         * Subscribe to state changes
         * @param {function} callback - Called with (updates) on state change
         * @returns {function} - Unsubscribe function
         */
        subscribe: function(callback) {
            var self = this;

            if (typeof callback !== 'function') {
                return function() {};
            }

            this._subscribers.push(callback);

            // Return unsubscribe function
            return function() {
                var index = self._subscribers.indexOf(callback);
                if (index > -1) {
                    self._subscribers.splice(index, 1);
                }
            };
        },

        // ====================================================================
        // Private: Notifications
        // ====================================================================

        _notifySubscribers: function(updates) {
            var self = this;
            this._subscribers.forEach(function(callback) {
                try {
                    callback(updates);
                } catch (e) {
                    console.error('[Store] Subscriber error:', e);
                }
            });
        },

        // ====================================================================
        // Polling Management
        // ====================================================================

        /**
         * Register a polling provider
         * Provider is a function that returns Promise<state updates>
         * @param {function} provider - Polling function
         */
        registerPollingProvider: function(provider) {
            if (typeof provider === 'function') {
                this._polling.providers.push(provider);
                console.log('[Store] Polling provider registered');
            }
        },

        /**
         * Start polling for state updates
         * @param {number} interval - Poll interval in ms (optional)
         */
        startPolling: function(interval) {
            var self = this;

            if (this._polling.active) {
                console.warn('[Store] Polling already active');
                return;
            }

            if (interval) {
                this._polling.interval = interval;
            }

            this._polling.active = true;
            console.log('[Store] Polling started (interval: ' + this._polling.interval + 'ms)');

            this._pollOnce();
        },

        /**
         * Stop polling
         */
        stopPolling: function() {
            this._polling.active = false;
            if (this._polling.timerHandle) {
                clearTimeout(this._polling.timerHandle);
                this._polling.timerHandle = null;
            }
            console.log('[Store] Polling stopped');
        },

        /**
         * Pause polling (can be resumed)
         */
        pausePolling: function() {
            this._polling.paused = true;
            console.log('[Store] Polling paused');
        },

        /**
         * Resume polling if paused
         */
        resumePolling: function() {
            var waspaused = this._polling.paused;
            this._polling.paused = false;
            if (waspaused && this._polling.active) {
                console.log('[Store] Polling resumed');
                this._pollOnce();
            }
        },

        // ====================================================================
        // Private: Polling Implementation
        // ====================================================================

        /**
         * Check if in critical state (should use fast polling)
         * @private
         */
        _isInCriticalState: function() {
            var alarmStatus = this.state.alarmState.state;
            var guestPending = this.state.guestState.requestPending;

            return CRITICAL_STATES.indexOf(alarmStatus) > -1 || guestPending;
        },

        /**
         * Update polling speed based on state
         * @private
         */
        _updatePollingSpeed: function() {
            if (!this._polling.active) {
                return;
            }

            var shouldBeFast = this._isInCriticalState();
            var currentFast = this._polling.interval === DEFAULT_FAST_POLL_INTERVAL;

            if (shouldBeFast && !currentFast) {
                this._polling.interval = DEFAULT_FAST_POLL_INTERVAL;
                console.log('[Store] Switched to fast polling (1s)');
                // Restart polling cycle immediately
                if (this._polling.timerHandle) {
                    clearTimeout(this._polling.timerHandle);
                }
                this._pollOnce();
            } else if (!shouldBeFast && currentFast) {
                this._polling.interval = DEFAULT_POLL_INTERVAL;
                console.log('[Store] Switched to normal polling (5s)');
            }
        },

        /**
         * Execute one polling cycle
         * @private
         */
        _pollOnce: function() {
            var self = this;

            if (!this._polling.active || this._polling.paused) {
                return;
            }

            // Execute all polling providers
            if (this._polling.providers.length > 0) {
                var promises = this._polling.providers.map(function(provider) {
                    return Promise.resolve(provider()).catch(function(e) {
                        console.error('[Store] Polling provider error:', e);
                        return null;
                    });
                });

                Promise.all(promises).then(function(results) {
                    // Merge results into state
                    results.forEach(function(result) {
                        if (result && typeof result === 'object') {
                            self.setState(result);
                        }
                    });

                    // Schedule next poll
                    self._polling.timerHandle = setTimeout(function() {
                        self._pollOnce();
                    }, self._polling.interval);
                });
            } else {
                // No providers, just schedule next
                this._polling.timerHandle = setTimeout(function() {
                    self._pollOnce();
                }, this._polling.interval);
            }
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.store = Store;

    // Add initial state to global state
    window.SmartDisplay.state = Store.getState();

    console.log('[SmartDisplay] Store registered');

})();
