/**
 * SmartDisplay First Boot Premium Intro
 * FAZ L5: Premium first-boot experience (2-3 seconds)
 * One-time only, graceful failure handling
 */

(function() {
    'use strict';

    // ========================================================================
    // Intro Module
    // ========================================================================
    var Intro = {
        // State
        _shown: false,
        _inProgress: false,
        _containerElement: null,
        _animationTimer: null,

        // ====================================================================
        // Public API
        // ====================================================================

        /**
         * Check if intro should be shown
         * @returns {boolean} - true if first boot and not yet shown
         */
        shouldShow: function() {
            var state = window.SmartDisplay.store.getState();
            return state.firstBoot === true && !this._shown;
        },

        /**
         * Play intro sequence
         * Returns promise that resolves when intro is complete or skipped
         * @returns {Promise<void>}
         */
        play: function() {
            var self = this;

            if (this._inProgress || this._shown) {
                return Promise.resolve();
            }

            this._inProgress = true;

            try {
                // Check for reduced motion preference
                var prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

                if (prefersReducedMotion) {
                    // Static intro (no animation)
                    return this._playStatic();
                } else {
                    // Animated intro with sound
                    return this._playAnimated();
                }
            } catch (e) {
                // Failure safety: skip intro immediately
                console.error('[Intro] Error during intro:', e);
                this._inProgress = false;
                this._shown = true;
                return Promise.resolve();
            }
        },

        /**
         * Skip intro immediately
         */
        skip: function() {
            this._cleanup();
            this._shown = true;
            this._inProgress = false;
            console.log('[Intro] Skipped');
        },

        // ====================================================================
        // Private: Playback Implementations
        // ====================================================================

        /**
         * Play animated intro sequence
         * @private
         */
        _playAnimated: function() {
            var self = this;

            return new Promise(function(resolve) {
                // Create intro container
                self._createContainer();

                // Fade in (300ms)
                self._containerElement.style.opacity = '0';
                self._containerElement.style.display = 'flex';
                
                // Trigger reflow
                void self._containerElement.offsetWidth;
                
                self._containerElement.style.transition = 'opacity 0.3s ease-in';
                self._containerElement.style.opacity = '1';

                // Hold (1.2 seconds)
                self._animationTimer = setTimeout(function() {
                    // Play subtle activation sound (very short, non-blocking)
                    self._playActivationSound();

                    // Soft glow (show for 1.5s)
                    setTimeout(function() {
                        // Fade out (300ms)
                        self._containerElement.style.transition = 'opacity 0.3s ease-out';
                        self._containerElement.style.opacity = '0';

                        // Cleanup and resolve
                        setTimeout(function() {
                            self._cleanup();
                            self._shown = true;
                            self._inProgress = false;
                            resolve();
                        }, 300);
                    }, 1500);
                }, 1200);
            });
        },

        /**
         * Play static intro (reduced motion)
         * @private
         */
        _playStatic: function() {
            var self = this;

            return new Promise(function(resolve) {
                // Create container
                self._createContainer();

                // Show static (no animation)
                self._containerElement.style.opacity = '0.85';
                self._containerElement.style.display = 'flex';
                self._containerElement.style.transition = 'none';

                // Hold for 2 seconds
                self._animationTimer = setTimeout(function() {
                    // Hide
                    self._containerElement.style.display = 'none';

                    self._cleanup();
                    self._shown = true;
                    self._inProgress = false;
                    resolve();
                }, 2000);
            });
        },

        /**
         * Create intro visual container
         * @private
         */
        _createContainer: function() {
            if (this._containerElement) {
                return;
            }

            var container = document.createElement('div');
            container.id = 'intro-container';
            container.className = 'intro-container';
            container.innerHTML = [
                '<div class="intro-content">',
                '  <div class="intro-logo-wrapper">',
                '    <div class="intro-glow"></div>',
                '    <div class="intro-logo">',
                '      <svg viewBox="0 0 100 100" width="80" height="80">',
                '        <circle cx="50" cy="50" r="45" fill="none" stroke="currentColor" stroke-width="2"/>',
                '        <text x="50" y="60" text-anchor="middle" font-size="32" fill="currentColor" font-weight="bold">SD</text>',
                '      </svg>',
                '    </div>',
                '  </div>',
                '  <h1 class="intro-title">SmartDisplay</h1>',
                '  <p class="intro-subtitle">System Ready</p>',
                '</div>'
            ].join('\n');

            document.body.appendChild(container);
            this._containerElement = container;

            console.log('[Intro] Container created');
        },

        /**
         * Play subtle activation sound
         * @private
         */
        _playActivationSound: function() {
            try {
                // Use Web Audio API for a subtle, generated beep
                var audioContext = new (window.AudioContext || window.webkitAudioContext)();
                
                // Very brief beep: 500ms at 800Hz
                var oscillator = audioContext.createOscillator();
                var gainNode = audioContext.createGain();

                oscillator.connect(gainNode);
                gainNode.connect(audioContext.destination);

                oscillator.frequency.value = 800;
                oscillator.type = 'sine';

                // Fade in and out (total 500ms)
                var now = audioContext.currentTime;
                gainNode.gain.setValueAtTime(0, now);
                gainNode.gain.linearRampToValueAtTime(0.1, now + 0.05);
                gainNode.gain.linearRampToValueAtTime(0, now + 0.5);

                oscillator.start(now);
                oscillator.stop(now + 0.5);

                console.log('[Intro] Activation sound played');
            } catch (e) {
                // Graceful failure - continue if audio fails
                console.warn('[Intro] Sound playback failed (continuing):', e.message);
            }
        },

        /**
         * Cleanup intro resources
         * @private
         */
        _cleanup: function() {
            if (this._animationTimer) {
                clearTimeout(this._animationTimer);
                this._animationTimer = null;
            }

            if (this._containerElement) {
                this._containerElement.remove();
                this._containerElement = null;
            }

            console.log('[Intro] Cleaned up');
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.intro = Intro;

    console.log('[SmartDisplay] Intro registered');

})();
