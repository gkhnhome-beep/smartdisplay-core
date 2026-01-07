/**
 * AlarmScene - Cinematic fullscreen alarm visualization
 * 
 * State-driven scene system that reacts to Alarmo state from Home Assistant
 * Displays appropriate visual urgency level without local alarm logic
 */

window.SmartDisplay = window.SmartDisplay || {};

window.SmartDisplay.AlarmScene = (function() {
    'use strict';

    var SCENE_TYPES = {
        NONE: 'none',
        COUNTDOWN: 'countdown-scene',
        EMERGENCY: 'emergency-scene',
        GUARDED: 'guarded-scene'
    };

    var scene = {
        currentScene: SCENE_TYPES.NONE,
        isInitialized: false,
        storeUnsubscribe: null,
        countdownUpdateTimer: null,

        /**
         * Initialize AlarmScene
         * Creates overlay container and subscribes to Store
         */
        init: function() {
            if (this.isInitialized) {
                console.log('[AlarmScene] Already initialized');
                return;
            }

            console.log('[AlarmScene] Initializing...');

            // Create overlay container
            var overlay = document.createElement('div');
            overlay.id = 'alarm-scene-overlay';
            overlay.className = 'alarm-scene-overlay';
            document.body.insertBefore(overlay, document.body.firstChild);

            // Subscribe to Store changes
            if (window.SmartDisplay.store) {
                this.storeUnsubscribe = window.SmartDisplay.store.subscribe(function() {
                    scene._onStoreChange();
                });
            }

            this.isInitialized = true;
            console.log('[AlarmScene] Mounted');

            // Initial render
            this._onStoreChange();
        },

        /**
         * Handle Store state changes
         * Determines appropriate scene based on Alarmo state
         */
        _onStoreChange: function() {
            var state = window.SmartDisplay.store.getState();
            
            if (!state) {
                this._setScene(SCENE_TYPES.NONE);
                return;
            }

            var alarmState = state.alarmoState || {};
            var authState = state.authState || {};
            var alarmoState = alarmState.alarmo_state || 'disarmed';
            var isTriggered = alarmState.alarmo_triggered === true;
            var delayRemaining = alarmState.delay_remaining || 0;

            // Scene priority: triggered > pending/arming. Do NOT block UI when simply armed.
            if (isTriggered) {
                this._showEmergencyScene(authState, delayRemaining);
            } else if (alarmoState === 'pending' || alarmoState === 'arming') {
                this._showCountdownScene(delayRemaining);
            } else {
                // Armed states keep UI fully usable; hide cinematic overlay
                this._setScene(SCENE_TYPES.NONE);
            }
        },

        /**
         * Display countdown scene (arming / pending)
         */
        _showCountdownScene: function(delayRemaining) {
            console.log('[AlarmScene] Scene changed: countdown (delay:', delayRemaining, 's)');

            var overlay = document.getElementById('alarm-scene-overlay');
            if (!overlay) return;

            // Set pulse speed based on time remaining
            var pulseSpeed = delayRemaining <= 10 ? '1.2s' : '3s';
            overlay.style.setProperty('--pulse-speed', pulseSpeed);

            // Generate HTML
            overlay.innerHTML = `
                <div class="countdown-content">
                    <div class="countdown-label">Alarm arming</div>
                    <div class="countdown-timer" id="alarm-countdown-value">${this._formatTime(delayRemaining)}</div>
                </div>
            `;

            var timerEl = document.getElementById('alarm-countdown-value');
            if (delayRemaining <= 10 && timerEl) {
                timerEl.classList.add('warning');
            }

            this._setScene(SCENE_TYPES.COUNTDOWN);

            // Update countdown without timers - wait for next Store update
        },

        /**
         * Display emergency scene (triggered)
         */
        _showEmergencyScene: function(authState, delayRemaining) {
            console.log('[AlarmScene] Scene changed: emergency (delay:', delayRemaining, 's)');

            var overlay = document.getElementById('alarm-scene-overlay');
            if (!overlay) return;

            var userRole = authState.role || 'guest';
            var isPinVisible = userRole === 'admin' || userRole === 'user';

            var html = `
                <div class="emergency-content">
                    <div class="emergency-title">ALARM TRIGGERED</div>
                    <div class="emergency-subtitle">Immediate action required</div>
            `;

            if (isPinVisible) {
                html += `
                    <div class="emergency-pin-section">
                        <input type="password" id="emergency-pin-input" class="emergency-pin-input" 
                               maxlength="6" placeholder="••••••" autocomplete="off">
                        
                        <div class="emergency-numpad">
                            <button class="emergency-numpad-btn" data-num="1">1</button>
                            <button class="emergency-numpad-btn" data-num="2">2</button>
                            <button class="emergency-numpad-btn" data-num="3">3</button>
                            <button class="emergency-numpad-btn" data-num="4">4</button>
                            <button class="emergency-numpad-btn" data-num="5">5</button>
                            <button class="emergency-numpad-btn" data-num="6">6</button>
                            <button class="emergency-numpad-btn" data-num="7">7</button>
                            <button class="emergency-numpad-btn" data-num="8">8</button>
                            <button class="emergency-numpad-btn" data-num="9">9</button>
                            <button class="emergency-numpad-btn" data-num="0">0</button>
                            <button class="emergency-numpad-btn clear" data-num="clear">⌫</button>
                        </div>

                        <button class="emergency-disarm-btn" id="emergency-disarm-btn">Disarm</button>
                    </div>
                `;
            } else {
                html += `
                    <div class="emergency-guest-warning">
                        <p>Guest mode active. Only administrators can disarm the alarm.</p>
                    </div>
                `;
            }

            if (delayRemaining > 0) {
                html += `
                    <div class="emergency-countdown">
                        <div class="emergency-countdown-label">Escalation in</div>
                        <div class="emergency-countdown-value" id="emergency-countdown-value">${this._formatTime(delayRemaining)}</div>
                    </div>
                `;
            }

            html += '</div>';

            overlay.innerHTML = html;

            // Setup event listeners
            if (isPinVisible) {
                this._setupEmergencyListeners();
            }

            this._setScene(SCENE_TYPES.EMERGENCY);

            // Trigger screen shake animation
            overlay.classList.add('shake');
            setTimeout(function() {
                overlay.classList.remove('shake');
            }, 400);
        },

        /**
         * Display guarded scene (armed_home / away / night)
         */
        _showGuardedScene: function(alarmoState) {
            console.log('[AlarmScene] Scene changed: guarded (' + alarmoState + ')');

            var overlay = document.getElementById('alarm-scene-overlay');
            if (!overlay) return;

            var modeLabel = {
                'armed_home': 'Home Armed',
                'armed_away': 'Away Mode',
                'armed_night': 'Night Mode'
            }[alarmoState] || 'Armed';

            overlay.innerHTML = `
                <div class="guarded-content">
                    <div class="guarded-icon"></div>
                    <div class="guarded-status">Protected</div>
                    <div class="guarded-mode">${modeLabel}</div>
                </div>
            `;

            this._setScene(SCENE_TYPES.GUARDED);
        },

        /**
         * Setup emergency PIN pad listeners
         */
        _setupEmergencyListeners: function() {
            var self = this;

            // Numpad buttons
            var numpadBtns = document.querySelectorAll('.emergency-numpad-btn');
            numpadBtns.forEach(function(btn) {
                btn.addEventListener('click', function() {
                    var num = this.getAttribute('data-num');
                    self._handleEmergencyInput(num);
                });
            });

            // Disarm button
            var disarmBtn = document.getElementById('emergency-disarm-btn');
            if (disarmBtn) {
                disarmBtn.addEventListener('click', function() {
                    self._submitEmergencyDisarm();
                });
            }

            // PIN input keyboard
            var pinInput = document.getElementById('emergency-pin-input');
            if (pinInput) {
                pinInput.addEventListener('keydown', function(e) {
                    if (e.key === 'Enter') {
                        e.preventDefault();
                        self._submitEmergencyDisarm();
                    } else if (e.key === 'Backspace') {
                        // Allow native backspace
                        return true;
                    } else if (e.key >= '0' && e.key <= '9') {
                        // Allow numbers
                        return true;
                    } else {
                        e.preventDefault();
                    }
                });

                pinInput.addEventListener('paste', function(e) {
                    e.preventDefault();
                });
            }
        },

        /**
         * Handle emergency numpad input
         */
        _handleEmergencyInput: function(num) {
            var pinInput = document.getElementById('emergency-pin-input');
            if (!pinInput) return;

            if (num === 'clear') {
                pinInput.value = '';
                console.log('[AlarmScene] PIN cleared');
            } else if (pinInput.value.length < 6) {
                pinInput.value += num;
            }
        },

        /**
         * Submit emergency disarm
         */
        _submitEmergencyDisarm: function() {
            var pinInput = document.getElementById('emergency-pin-input');
            if (!pinInput) return;

            var code = pinInput.value.trim();

            if (!code || code.length < 4) {
                console.log('[AlarmScene] PIN too short or empty');
                return;
            }

            console.log('[AlarmScene] Submitting emergency disarm, code length:', code.length);

            // Call API to disarm
            if (window.SmartDisplay.api && window.SmartDisplay.api.client) {
                window.SmartDisplay.api.client.post('/ui/alarmo/disarm', {
                    headers: {
                        'X-User-Role': (window.SmartDisplay.store.getState().authState || {}).role || 'guest'
                    }
                }, { code: code })
                    .then(function(envelope) {
                        console.log('[AlarmScene] Disarm response:', envelope);
                        var response = envelope.response || {};
                        if (!response.ok) {
                            throw new Error(response.error || 'Disarm failed');
                        }
                        console.log('[AlarmScene] Alarm disarmed successfully');
                        // Scene will update via Store change
                    })
                    .catch(function(err) {
                        console.error('[AlarmScene] Disarm error:', err);
                        pinInput.value = '';
                    });
            }
        },

        /**
         * Set active scene
         */
        _setScene: function(newScene) {
            if (this.currentScene === newScene) return;

            var overlay = document.getElementById('alarm-scene-overlay');
            if (!overlay) return;

            // Remove all scene classes
            Object.values(SCENE_TYPES).forEach(function(sceneType) {
                overlay.classList.remove(sceneType);
            });

            // Set new scene
            this.currentScene = newScene;
            if (newScene !== SCENE_TYPES.NONE) {
                overlay.classList.add(newScene);
                overlay.classList.add('active');
            } else {
                overlay.classList.remove('active');
            }
        },

        /**
         * Format time as MM:SS or SS
         */
        _formatTime: function(seconds) {
            if (seconds < 0) return '00';
            if (seconds < 60) return String(seconds).padStart(2, '0');
            
            var mins = Math.floor(seconds / 60);
            var secs = seconds % 60;
            return mins + ':' + String(secs).padStart(2, '0');
        },

        /**
         * Cleanup
         */
        destroy: function() {
            console.log('[AlarmScene] Destroying...');

            if (this.storeUnsubscribe) {
                this.storeUnsubscribe();
                this.storeUnsubscribe = null;
            }

            if (this.countdownUpdateTimer) {
                clearInterval(this.countdownUpdateTimer);
                this.countdownUpdateTimer = null;
            }

            var overlay = document.getElementById('alarm-scene-overlay');
            if (overlay) {
                overlay.remove();
            }

            this.currentScene = SCENE_TYPES.NONE;
            this.isInitialized = false;
        }
    };

    return scene;
})();
