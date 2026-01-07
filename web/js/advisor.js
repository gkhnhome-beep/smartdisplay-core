/**
 * SmartDisplay Admin AI Advisor
 * FAZ L4: Silent, contextual hints for admin users
 * Never chatty, never blocking, graceful degradation
 */

(function() {
    'use strict';

    // ========================================================================
    // Advisor Module
    // ========================================================================
    var Advisor = {
        // UI element
        _bubbleElement: null,
        _dismissTimer: null,
        _lastHintId: null,

        // ====================================================================
        // Context Detection & Hint Generation
        // ====================================================================

        /**
         * Determine if a hint should be shown based on context
         * Pure function - no side effects
         * @param {object} context - Current app context
         * @returns {object|null} - { id, text } or null
         */
        getHint: function(context) {
            if (!context) {
                return null;
            }

            // Only show hints to admins
            if (context.role !== 'admin') {
                return null;
            }

            // Don't show if alarm is triggered (lockdown mode)
            if (context.alarmState === 'triggered') {
                return null;
            }

            // Check for HA not connected
            if (!context.haIsConnected && context.haIsConfigured) {
                return {
                    id: 'ha-disconnected',
                    text: 'Home Assistant connection lost. Check settings.'
                };
            }

            // Check for HA configured but sync not done
            if (context.haIsConnected && context.haIsConfigured && !context.haSyncDone) {
                return {
                    id: 'ha-sync-pending',
                    text: 'Initial sync pending. Complete in settings.'
                };
            }

            // Check for guest session active too long (> 60 minutes)
            if (context.guestIsActive && context.guestApprovedAt) {
                var now = Date.now();
                var guestDuration = now - context.guestApprovedAt;
                var ONE_HOUR = 60 * 60 * 1000;

                if (guestDuration > ONE_HOUR) {
                    return {
                        id: 'guest-duration',
                        text: 'Guest session has been active for over an hour.'
                    };
                }
            }

            // Check for admin in settings (helpful context)
            if (context.currentView === 'settings') {
                return {
                    id: 'settings-context',
                    text: 'Home Assistant connection status and sync shown above.'
                };
            }

            return null;
        },

        // ====================================================================
        // Public API
        // ====================================================================

        /**
         * Initialize advisor UI (create bubble element)
         */
        init: function() {
            if (this._bubbleElement) {
                return; // Already initialized
            }

            // Create bubble element
            this._bubbleElement = document.createElement('div');
            this._bubbleElement.id = 'advisor-bubble';
            this._bubbleElement.className = 'advisor-bubble';
            this._bubbleElement.style.display = 'none';
            this._bubbleElement.innerHTML = '<div class="advisor-bubble-text"></div>';

            document.body.appendChild(this._bubbleElement);

            console.log('[Advisor] Initialized');
        },

        /**
         * Check context and show hint if applicable
         * Called from controllers when context changes
         * @param {object} context - Current app state
         */
        checkAndShow: function(context) {
            if (!this._bubbleElement) {
                this.init();
            }

            // Clear any pending dismiss timer
            if (this._dismissTimer) {
                clearTimeout(this._dismissTimer);
                this._dismissTimer = null;
            }

            var hint = this.getHint(context);

            // No hint or same hint as last time - hide
            if (!hint || (this._lastHintId && hint.id === this._lastHintId)) {
                this.dismiss();
                return;
            }

            // Show new hint
            this._lastHintId = hint.id;
            this._showBubble(hint.text);

            // Auto-dismiss after 6 seconds
            var self = this;
            this._dismissTimer = setTimeout(function() {
                self.dismiss();
            }, 6000);

            // Update store
            window.SmartDisplay.store.setState({
                aiAdvisorState: {
                    lastHintAt: Date.now(),
                    currentHint: hint
                }
            });
        },

        /**
         * Dismiss current hint
         */
        dismiss: function() {
            if (this._dismissTimer) {
                clearTimeout(this._dismissTimer);
                this._dismissTimer = null;
            }

            if (this._bubbleElement) {
                this._bubbleElement.style.display = 'none';
            }

            this._lastHintId = null;

            // Update store - use undefined instead of null for proper merging
            window.SmartDisplay.store.setState({
                aiAdvisorState: {
                    currentHint: undefined
                }
            });
        },

        /**
         * Manually trigger a hint (for testing)
         * @param {string} text - Hint text
         */
        showManual: function(text) {
            if (!text) return;

            if (!this._bubbleElement) {
                this.init();
            }

            this._lastHintId = 'manual-' + Date.now();
            this._showBubble(text);

            // Auto-dismiss after 6 seconds
            var self = this;
            if (this._dismissTimer) {
                clearTimeout(this._dismissTimer);
            }
            this._dismissTimer = setTimeout(function() {
                self.dismiss();
            }, 6000);
        },

        // ====================================================================
        // Private: UI Rendering
        // ====================================================================

        /**
         * Show bubble with text
         * @private
         */
        _showBubble: function(text) {
            if (!this._bubbleElement) {
                return;
            }

            // Check for reduced motion preference
            var prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

            var textEl = this._bubbleElement.querySelector('.advisor-bubble-text');
            if (textEl) {
                textEl.textContent = text;
            }

            this._bubbleElement.style.display = 'block';

            // Fade in (no animation if reduced motion)
            if (!prefersReducedMotion) {
                this._bubbleElement.style.opacity = '0';
                // Trigger reflow to force animation
                void this._bubbleElement.offsetWidth;
                this._bubbleElement.style.transition = 'opacity 0.3s ease-in';
                this._bubbleElement.style.opacity = '0.85';
            } else {
                this._bubbleElement.style.opacity = '0.85';
                this._bubbleElement.style.transition = 'none';
            }
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.advisor = Advisor;

    console.log('[SmartDisplay] Advisor registered');

})();
