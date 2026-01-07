/**
 * SmartDisplay Admin Trace UX
 * FAZ L6: Observable, reversible admin actions
 * Not an audit log, just recent action feedback
 */

(function() {
    'use strict';

    // ========================================================================
    // Trace Module
    // ========================================================================
    var Trace = {
        // UI element
        _containerElement: null,
        _maxEntries: 5,

        // ====================================================================
        // Public API
        // ====================================================================

        /**
         * Initialize trace UI
         */
        init: function() {
            if (this._containerElement) {
                return; // Already initialized
            }

            // Create container
            this._containerElement = document.createElement('div');
            this._containerElement.id = 'admin-trace';
            this._containerElement.className = 'admin-trace';
            this._containerElement.style.display = 'none';

            document.body.appendChild(this._containerElement);

            console.log('[Trace] Initialized');
        },

        /**
         * Add a trace entry (admin action)
         * Only call for actual admin actions - no guest/user actions
         * @param {string} label - Human-readable action description
         */
        add: function(label) {
            if (!label || typeof label !== 'string') {
                return;
            }

            if (!this._containerElement) {
                this.init();
            }

            var state = window.SmartDisplay.store.getState();

            // Only show trace in admin mode
            if (state.authState.role !== 'admin') {
                return;
            }

            var entry = {
                label: label,
                timestamp: Date.now()
            };

            // Update store
            var recent = state.adminTrace.recent || [];
            recent.unshift(entry); // Add to front

            // Keep only max entries
            if (recent.length > this._maxEntries) {
                recent = recent.slice(0, this._maxEntries);
            }

            window.SmartDisplay.store.setState({
                adminTrace: {
                    recent: recent
                }
            });

            // Render
            this._render(recent);

            console.log('[Trace] Entry added:', label);
        },

        /**
         * Clear all trace entries
         */
        clear: function() {
            window.SmartDisplay.store.setState({
                adminTrace: {
                    recent: []
                }
            });

            if (this._containerElement) {
                this._containerElement.innerHTML = '';
                this._containerElement.style.display = 'none';
            }

            console.log('[Trace] Cleared');
        },

        /**
         * Show trace UI
         */
        show: function() {
            if (!this._containerElement) {
                this.init();
            }

            this._containerElement.style.display = 'flex';
            console.log('[Trace] Shown');
        },

        /**
         * Hide trace UI
         */
        hide: function() {
            if (this._containerElement) {
                this._containerElement.style.display = 'none';
            }

            console.log('[Trace] Hidden');
        },

        // ====================================================================
        // Private: UI Rendering
        // ====================================================================

        /**
         * Render trace entries
         * @private
         */
        _render: function(entries) {
            if (!this._containerElement || !entries || entries.length === 0) {
                if (this._containerElement) {
                    this._containerElement.style.display = 'none';
                }
                return;
            }

            var self = this;
            var html = [];

            html.push('<div class="trace-stack">');

            entries.forEach(function(entry, index) {
                var age = Date.now() - entry.timestamp;
                var opacity = 1 - (index * 0.2); // Older entries fade

                html.push(
                    '<div class="trace-entry" style="opacity: ' + Math.max(0.4, opacity) + '">',
                    '  <span class="trace-label">' + self._escapeHtml(entry.label) + '</span>',
                    '  <span class="trace-time">' + self._formatTime(age) + '</span>',
                    '</div>'
                );
            });

            html.push('</div>');

            this._containerElement.innerHTML = html.join('\n');
            this._containerElement.style.display = 'flex';
        },

        /**
         * Format time elapsed in human-readable form
         * @private
         */
        _formatTime: function(ms) {
            var seconds = Math.floor(ms / 1000);
            var minutes = Math.floor(seconds / 60);

            if (minutes > 0) {
                return minutes + 'm ago';
            }

            if (seconds > 0) {
                return seconds + 's ago';
            }

            return 'just now';
        },

        /**
         * Escape HTML to prevent XSS
         * @private
         */
        _escapeHtml: function(text) {
            var map = {
                '&': '&amp;',
                '<': '&lt;',
                '>': '&gt;',
                '"': '&quot;',
                "'": '&#039;'
            };
            return text.replace(/[&<>"']/g, function(m) {
                return map[m];
            });
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.trace = Trace;

    console.log('[SmartDisplay] Trace registered');

})();
