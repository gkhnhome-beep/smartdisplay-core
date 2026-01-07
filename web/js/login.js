/**
 * SmartDisplay Login Controller
 * FAZ L1: PIN-based authentication
 */

(function() {
    'use strict';

    // ========================================================================
    // Login Controller
    // ========================================================================
    var LoginController = {
        pin: '',
        maxPinLength: 4,
        isValidating: false,
        error: false,

        /**
         * Initialize login controller
         */
        init: function() {
            console.log('[Login] Initializing controller');
            this.pin = '';
            this.error = false;
            this.isValidating = false;
        },

        /**
         * Handle number button click
         * @param {string} digit - Digit pressed (0-9)
         */
        onDigitPress: function(digit) {
            if (this.isValidating) return;
            if (this.pin.length >= this.maxPinLength) return;

            this.pin += digit;
            this.error = false;
            console.log('[Login] PIN length: ' + this.pin.length);

            // Auto-submit when PIN length reached
            if (this.pin.length === this.maxPinLength) {
                this.submitPIN();
            }
        },

        /**
         * Handle backspace
         */
        onBackspace: function() {
            if (this.isValidating) return;
            this.pin = this.pin.slice(0, -1);
            this.error = false;
            console.log('[Login] PIN length: ' + this.pin.length);
        },

        /**
         * Clear PIN input
         */
        clear: function() {
            this.pin = '';
            this.error = false;
            this.isValidating = false;
        },

        /**
         * Submit PIN for validation
         */
        submitPIN: function() {
            var self = this;

            if (this.pin.length !== this.maxPinLength) {
                return;
            }

            this.isValidating = true;
            console.log('[Login] Validating PIN...');

            // Store PIN in auth state temporarily (memory only)
            window.SmartDisplay.store.setState({
                authState: {
                    authenticated: false,
                    role: 'guest',
                    pin: this.pin
                }
            });

            // Call backend to validate PIN
            // Using /ui/home/state as test endpoint (middleware will validate PIN)
            window.SmartDisplay.api.client.get('/ui/home/state')
                .then(function(response) {
                    // Backend validated PIN successfully
                    // Extract role from auth context (backend should include it)
                    console.log('[Login] PIN validated successfully');
                    self.onLoginSuccess();
                })
                .catch(function(err) {
                    console.error('[Login] PIN validation failed:', err);
                    self.onLoginFailure();
                });
        },

        /**
         * Handle successful login
         */
        onLoginSuccess: function() {
            var self = this;

            console.log('[Login] Login successful');

            // Get current auth state (PIN was already set)
            var currentAuthState = window.SmartDisplay.store.state.authState;

            // Determine role from backend response
            // Backend auth middleware validated PIN and set role in context
            // For now, we'll check the PIN against known PINs
            var role = this.determineRole(this.pin);

            // Update store with authenticated state
            window.SmartDisplay.store.setState({
                authState: {
                    authenticated: true,
                    role: role,
                    pin: this.pin  // Keep in memory only
                },
                currentRole: role  // Backward compatibility
            });

            // Clear PIN from UI
            this.clear();

            // Fetch HA status after successful admin login
            if (role === 'admin' && window.SmartDisplay.settings) {
                window.SmartDisplay.settings.fetchHAStatus()
                    .catch(function(err) {
                        console.error('[Login] Failed to fetch HA status after login:', err);
                    });
            }

            // Route to home view
            setTimeout(function() {
                if (window.SmartDisplay.viewManager) {
                    window.SmartDisplay.viewManager.routeToView('home');
                }
            }, 100);
        },

        /**
         * Handle login failure
         */
        onLoginFailure: function() {
            console.log('[Login] Login failed');

            this.isValidating = false;
            this.error = true;

            // Clear PIN
            this.pin = '';

            // Reset store auth state to guest
            window.SmartDisplay.store.setState({
                authState: {
                    authenticated: false,
                    role: 'guest',
                    pin: null
                }
            });

            // Show error feedback briefly (visual shake handled by view)
            var self = this;
            setTimeout(function() {
                self.error = false;
            }, 2000);
        },

        /**
         * Determine role from PIN
         * FAZ L1: Simple PIN-based mapping
         * TODO: Backend should return role in response
         * @param {string} pin
         * @returns {string} - Role
         */
        determineRole: function(pin) {
            // Match backend PIN configuration
            if (pin === '1234') return 'admin';
            if (pin === '5678') return 'user';
            return 'guest';
        },

        /**
         * Get masked PIN display
         * @returns {string} - Masked PIN (●●●●)
         */
        getMaskedPIN: function() {
            var masked = '';
            for (var i = 0; i < this.pin.length; i++) {
                masked += '●';
            }
            return masked;
        },

        /**
         * Logout (clears auth state)
         */
        logout: function() {
            console.log('[Login] Logging out');

            // Clear store
            window.SmartDisplay.store.setState({
                authState: {
                    authenticated: false,
                    role: 'guest',
                    pin: null
                },
                currentRole: 'guest'
            });

            // Clear local state
            this.clear();

            // Route to login view
            if (window.SmartDisplay.viewManager) {
                window.SmartDisplay.viewManager.routeToView('login');
            }
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.loginController = LoginController;

    console.log('[SmartDisplay] Login controller registered');

})();
