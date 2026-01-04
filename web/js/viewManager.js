/**
 * SmartDisplay View Manager
 * Handles view routing and lifecycle based on application state
 */

(function() {
    'use strict';

    // ========================================================================
    // View Definitions
    // ========================================================================

    /**
     * FirstBoot View
     * Shown on initial setup/first boot
     * Backend-driven step navigation
     */
    var FirstBootView = {
        id: 'first-boot',
        name: 'FirstBoot',

        mount: function() {
            console.log('[View] Mounting FirstBootView');
            var container = document.getElementById('app');
            
            container.innerHTML = '';
            var viewElement = document.createElement('div');
            viewElement.id = this.id;
            viewElement.className = 'view view-first-boot';
            
            viewElement.innerHTML = [
                '<div class="view-content">',
                '  <div class="first-boot-container">',
                '    <div class="first-boot-progress" id="fb-progress">--</div>',
                '    <div class="first-boot-step-number" id="fb-step-number">--</div>',
                '    <div class="first-boot-message" id="fb-message">Loading...</div>',
                '    <div class="first-boot-description" id="fb-description">--</div>',
                '    <div class="first-boot-error" id="fb-error" style="display:none;"></div>',
                '    <div class="first-boot-controls">',
                '      <button class="fb-btn fb-btn-back" id="fb-btn-back" disabled>Back</button>',
                '      <button class="fb-btn fb-btn-next" id="fb-btn-next" disabled>Next</button>',
                '      <button class="fb-btn fb-btn-complete" id="fb-btn-complete" style="display:none;" disabled>Complete</button>',
                '    </div>',
                '  </div>',
                '</div>'
            ].join('\n');
            
            container.appendChild(viewElement);
            
            // Setup event listeners
            this._setupEventListeners();
            
            // Load initial step
            this._loadStep();
        },

        unmount: function() {
            console.log('[View] Unmounting FirstBootView');
            var element = document.getElementById(this.id);
            if (element) {
                element.remove();
            }
        },

        update: function(data) {
            console.log('[View] Updating FirstBootView', data);
            // Update is handled by controller callbacks
        },

        // ====================================================================
        // Private Methods
        // ====================================================================

        _setupEventListeners: function() {
            var self = this;

            var backBtn = document.getElementById('fb-btn-back');
            var nextBtn = document.getElementById('fb-btn-next');
            var completeBtn = document.getElementById('fb-btn-complete');

            if (backBtn) {
                backBtn.addEventListener('click', function() {
                    self._handleBack();
                });
            }

            if (nextBtn) {
                nextBtn.addEventListener('click', function() {
                    self._handleNext();
                });
            }

            if (completeBtn) {
                completeBtn.addEventListener('click', function() {
                    self._handleComplete();
                });
            }
        },

        _loadStep: function() {
            var self = this;

            if (!window.SmartDisplay.firstBootController) {
                console.error('[FirstBootView] First Boot controller not initialized');
                this._showError('Controller not initialized');
                return;
            }

            // Initialize controller if not done
            if (!window.SmartDisplay.firstBootController.currentStep) {
                window.SmartDisplay.firstBootController.init()
                    .then(function() {
                        self._renderStep();
                    })
                    .catch(function(err) {
                        self._showError('Failed to load first boot');
                    });
            } else {
                this._renderStep();
            }
        },

        _renderStep: function() {
            var controller = window.SmartDisplay.firstBootController;
            var step = controller.currentStep;

            if (!step) {
                this._showError('No step data available');
                return;
            }

            console.log('[FirstBootView] Rendering step:', step);

            // Render step details
            var progressEl = document.getElementById('fb-progress');
            var stepNumEl = document.getElementById('fb-step-number');
            var messageEl = document.getElementById('fb-message');
            var descriptionEl = document.getElementById('fb-description');
            var backBtn = document.getElementById('fb-btn-back');
            var nextBtn = document.getElementById('fb-btn-next');
            var completeBtn = document.getElementById('fb-btn-complete');

            // Update progress indicator
            if (progressEl && step.progress) {
                progressEl.textContent = 'Step ' + step.currentStep + ' of ' + step.totalSteps;
            }

            // Update step number
            if (stepNumEl && step.currentStep) {
                stepNumEl.textContent = step.currentStep;
            }

            // Update message
            if (messageEl && step.message) {
                messageEl.textContent = step.message;
            }

            // Update description
            if (descriptionEl && step.description) {
                descriptionEl.textContent = step.description;
            }

            // Update button states
            if (backBtn) {
                backBtn.disabled = controller.isLoading || !step.showBack;
            }

            if (nextBtn) {
                nextBtn.disabled = controller.isLoading || !step.showNext;
                if (step.showNext) {
                    nextBtn.style.display = 'inline-block';
                } else {
                    nextBtn.style.display = 'none';
                }
            }

            if (completeBtn) {
                completeBtn.disabled = controller.isLoading || !step.showComplete;
                if (step.showComplete) {
                    completeBtn.style.display = 'inline-block';
                } else {
                    completeBtn.style.display = 'none';
                }
            }

            // Clear error if previously shown
            this._clearError();
        },

        _handleNext: function() {
            var self = this;
            var controller = window.SmartDisplay.firstBootController;

            console.log('[FirstBootView] Next button clicked');

            this._setButtonsDisabled(true);

            controller.nextStep()
                .then(function() {
                    self._renderStep();
                })
                .catch(function(err) {
                    self._showError(err.message || 'Failed to move to next step');
                    self._setButtonsDisabled(false);
                });
        },

        _handleBack: function() {
            var self = this;
            var controller = window.SmartDisplay.firstBootController;

            console.log('[FirstBootView] Back button clicked');

            this._setButtonsDisabled(true);

            controller.previousStep()
                .then(function() {
                    self._renderStep();
                })
                .catch(function(err) {
                    self._showError(err.message || 'Failed to move to previous step');
                    self._setButtonsDisabled(false);
                });
        },

        _handleComplete: function() {
            var self = this;
            var controller = window.SmartDisplay.firstBootController;

            console.log('[FirstBootView] Complete button clicked');

            this._setButtonsDisabled(true);

            controller.complete()
                .then(function() {
                    console.log('[FirstBootView] First boot complete, triggering route change');
                    // Store will be updated to firstBoot: false, triggering route change
                })
                .catch(function(err) {
                    self._showError(err.message || 'Failed to complete first boot');
                    self._setButtonsDisabled(false);
                });
        },

        _showError: function(message) {
            var errorEl = document.getElementById('fb-error');
            if (errorEl) {
                errorEl.textContent = 'Error: ' + message;
                errorEl.style.display = 'block';
            }
        },

        _clearError: function() {
            var errorEl = document.getElementById('fb-error');
            if (errorEl) {
                errorEl.style.display = 'none';
                errorEl.textContent = '';
            }
        },

        _setButtonsDisabled: function(disabled) {
            var backBtn = document.getElementById('fb-btn-back');
            var nextBtn = document.getElementById('fb-btn-next');
            var completeBtn = document.getElementById('fb-btn-complete');

            if (backBtn) backBtn.disabled = disabled;
            if (nextBtn) nextBtn.disabled = disabled;
            if (completeBtn) completeBtn.disabled = disabled;
        }
    };

    /**
     * Home View
     * Main dashboard with AI insight, quick tiles, status indicators
     */
    var HomeView = {
        id: 'home',
        name: 'Home',
        inactivityTimeout: null,
        inactivityDelay: 30000, // 30 seconds
        clockUpdateInterval: null,

        mount: function() {
            console.log('[View] Mounting HomeView');
            var container = document.getElementById('app');
            
            container.innerHTML = '';
            var viewElement = document.createElement('div');
            viewElement.id = this.id;
            viewElement.className = 'view view-home';
            
            // Calm layout structure with clickable header
            viewElement.innerHTML = [
                '<div class="home-header" id="home-header"></div>',
                '<div class="home-container">',
                '  <div class="home-idle-screen" id="home-idle-screen">',
                '    <div class="idle-clock" id="idle-clock">--:--</div>',
                '    <div class="idle-date" id="idle-date">--</div>',
                '    <div class="idle-alarm" id="idle-alarm">--</div>',
                '    <div class="idle-ai" id="idle-ai">--</div>',
                '  </div>',
                '  <div class="home-active-screen" id="home-active-screen" style="display:none;">',
                '    <div class="active-status-bar">',
                '      <div class="status-datetime" id="status-datetime">--:--</div>',
                '      <div class="status-indicators">',
                '        <span class="status-ha" id="status-ha">HA: --</span>',
                '        <span class="status-alarm" id="status-alarm">--</span>',
                '      </div>',
                '    </div>',
                '    <div class="active-content">',
                '      <div class="ai-card" id="ai-card">',
                '        <div class="ai-insight" id="ai-insight">--</div>',
                '        <div class="ai-severity" id="ai-severity">--</div>',
                '      </div>',
                '    </div>',
                '  </div>',
                '</div>'
            ].join('\n');
            
            container.appendChild(viewElement);
            
            // Setup event listeners (must happen after DOM is created)
            this._setupEventListeners();
            
            // Initialize controller
            this._initController();
            
            // Start clock updates
            this._startClockUpdates();
        },

        unmount: function() {
            console.log('[View] Unmounting HomeView');
            
            // Cleanup timers
            if (this.inactivityTimeout) {
                clearTimeout(this.inactivityTimeout);
                this.inactivityTimeout = null;
            }
            
            if (this.clockUpdateInterval) {
                clearInterval(this.clockUpdateInterval);
                this.clockUpdateInterval = null;
            }
            
            var element = document.getElementById(this.id);
            if (element) {
                element.remove();
            }
        },

        update: function(data) {
            console.log('[View] Updating HomeView', data);
            
            var homeState = data.homeState || {};
            
            // Update AI card
            if (homeState.aiInsight !== undefined) {
                var aiInsightEl = document.getElementById('ai-insight');
                if (aiInsightEl) {
                    aiInsightEl.textContent = homeState.aiInsight || '--';
                }
            }
            
            if (homeState.aiSeverity !== undefined) {
                var aiSeverityEl = document.getElementById('ai-severity');
                if (aiSeverityEl) {
                    aiSeverityEl.textContent = homeState.aiSeverity || '--';
                }
            }
            
            // Update idle AI one-liner
            if (homeState.aiOneLiner !== undefined) {
                var idleAiEl = document.getElementById('idle-ai');
                if (idleAiEl) {
                    idleAiEl.textContent = homeState.aiOneLiner || '';
                }
            }
            
            // Update temperature
            if (homeState.temperature !== undefined) {
                var tempDisplay = homeState.temperature || '--';
                if (typeof homeState.temperature === 'number') {
                    tempDisplay = homeState.temperature.toFixed(1) + '°';
                }
                
                var activeTemp = document.getElementById('active-temp');
                if (activeTemp) {
                    activeTemp.textContent = tempDisplay;
                }
            }
            
            // Update alarm status
            if (data.alarmState) {
                var alarmStateLabel = this._formatAlarmLabel(data.alarmState);
                var alarmStatusEl = document.getElementById('status-alarm');
                if (alarmStatusEl) {
                    alarmStatusEl.textContent = alarmStateLabel;
                }
                
                var idleAlarmEl = document.getElementById('idle-alarm');
                if (idleAlarmEl) {
                    idleAlarmEl.textContent = 'Alarm: ' + alarmStateLabel;
                }
            }
        },

        // ====================================================================
        // Private Methods
        // ====================================================================

        _setupEventListeners: function() {
            var self = this;
            var container = document.getElementById(this.id);

            if (!container) return;

            // Header click - toggle menu
            var header = document.getElementById('home-header');
            if (header) {
                header.addEventListener('click', function() {
                    console.log('[HomeView] Header clicked - toggling menu');
                    var overlay = document.getElementById('menu-overlay');
                    if (overlay && overlay.classList.contains('menu-open')) {
                        window.SmartDisplay.viewManager.closeMenu();
                    } else {
                        window.SmartDisplay.viewManager.openMenu();
                    }
                });
            }

            // Tap anywhere (except header) to activate
            container.addEventListener('click', function(e) {
                // Don't trigger if header was clicked
                if (!e.target.closest('#home-header')) {
                    self._handleTap();
                }
            });

            container.addEventListener('touchstart', function(e) {
                // Don't trigger if header was touched
                if (!e.target.closest('#home-header')) {
                    self._handleTap();
                }
            });

            // Listen for active/inactive events
            document.addEventListener('home-active', function() {
                self._showActiveScreen();
            });

            document.addEventListener('home-inactive', function() {
                self._showIdleScreen();
            });
        },

        _initController: function() {
            var self = this;

            if (!window.SmartDisplay.homeController) {
                console.error('[HomeView] Home controller not initialized');
                return;
            }

            // Initialize if not done
            if (!window.SmartDisplay.homeController.currentState) {
                window.SmartDisplay.homeController.init()
                    .catch(function(err) {
                        console.error('[HomeView] Failed to init controller:', err);
                    });
            }
        },

        _startClockUpdates: function() {
            var self = this;

            // Update clock immediately
            this._updateClock();

            // Update clock every second
            this.clockUpdateInterval = setInterval(function() {
                self._updateClock();
            }, 1000);
        },

        _updateClock: function() {
            var now = new Date();
            
            // Format time (HH:MM)
            var hours = String(now.getHours()).padStart(2, '0');
            var minutes = String(now.getMinutes()).padStart(2, '0');
            var timeStr = hours + ':' + minutes;

            // Format date (Day, Month DD)
            var dayNames = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
            var monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
                            'July', 'August', 'September', 'October', 'November', 'December'];
            var dayStr = dayNames[now.getDay()] + ', ' + monthNames[now.getMonth()] + ' ' + now.getDate();

            // Update idle clock
            var idleClockEl = document.getElementById('idle-clock');
            if (idleClockEl) {
                idleClockEl.textContent = timeStr;
            }

            var idleDateEl = document.getElementById('idle-date');
            if (idleDateEl) {
                idleDateEl.textContent = dayStr;
            }

            // Update active status bar clock
            var statusDatetimeEl = document.getElementById('status-datetime');
            if (statusDatetimeEl) {
                statusDatetimeEl.textContent = timeStr;
            }
        },

        _handleTap: function() {
            var controller = window.SmartDisplay.homeController;
            if (!controller) return;

            console.log('[HomeView] Tap detected');

            // Mark as active
            controller.setActive();

            // Reset inactivity timer
            if (this.inactivityTimeout) {
                clearTimeout(this.inactivityTimeout);
            }

            // Check for reduced motion preference
            var prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

            if (!prefersReducedMotion) {
                // Set inactivity timeout (only if not reduced motion)
                this.inactivityTimeout = setTimeout(function() {
                    controller.setInactive();
                }, this.inactivityDelay);
            }
        },

        _formatAlarmLabel: function(alarmState) {
            if (!alarmState) {
                return '--';
            }

            var state = (alarmState.state || 'unknown').toLowerCase();

            if (alarmState.triggered || state === 'triggered') {
                return 'Triggered';
            }

            if (state === 'disarmed') {
                return 'Disarmed';
            }

            if (state.startsWith('armed_')) {
                var mode = state.split('_')[1];
                return 'Armed (' + mode.charAt(0).toUpperCase() + mode.slice(1) + ')';
            }

            if (state === 'pending') {
                return 'Pending';
            }

            if (state === 'arming') {
                return 'Arming';
            }

            return state.charAt(0).toUpperCase() + state.slice(1);
        },

        _showIdleScreen: function() {
            console.log('[HomeView] Showing idle screen');
            
            var idleScreen = document.getElementById('home-idle-screen');
            var activeScreen = document.getElementById('home-active-screen');

            if (idleScreen) idleScreen.style.display = 'flex';
            if (activeScreen) activeScreen.style.display = 'none';
        },

        _showActiveScreen: function() {
            console.log('[HomeView] Showing active screen');
            
            var idleScreen = document.getElementById('home-idle-screen');
            var activeScreen = document.getElementById('home-active-screen');

            if (idleScreen) idleScreen.style.display = 'none';
            if (activeScreen) activeScreen.style.display = 'block';
        }
    };

    /**
     * Alarm View
     * Alarm control and status - backend-driven UI
     */
    var AlarmView = {
        id: 'alarm',
        name: 'Alarm',
        _listenersSetup: false,

        mount: function() {
            console.log('[View] Mounting AlarmView');
            var container = document.getElementById('app');

            if (!container) {
                return;
            }

            container.innerHTML = '';
            var viewElement = document.createElement('div');
            viewElement.id = this.id;
            viewElement.className = 'view view-alarm';

            viewElement.innerHTML = [
                '<div class="alarm-container">',
                '  <div class="alarm-header" id="alarm-header">',
                '    <div class="alarm-mode" id="alarm-mode">--</div>',
                '    <div class="alarm-message" id="alarm-message">Waiting for alarm data...</div>',
                '    <div class="alarm-context" id="alarm-context">Waiting...</div>',
                '  </div>',
                '  <div class="alarm-countdown" id="alarm-countdown-container" style="display:none;">',
                '    <div class="countdown-label" id="countdown-label">--</div>',
                '    <div class="countdown-value" id="countdown-value">--</div>',
                '  </div>',
                '  <div class="alarm-connection-warning" id="alarm-connection-warning" style="display:none;">Connection issue. Retrying\u2026</div>',
                '  <div class="alarm-actions" id="alarm-actions" style="display:none;">',
                '    <div class="action-status" id="action-status" style="display:none;"></div>',
                '    <div class="action-buttons">',
                '      <button class="alarm-action-btn" id="btn-arm-home" disabled>Arm Home</button>',
                '      <button class="alarm-action-btn" id="btn-arm-away" disabled>Arm Away</button>',
                '      <button class="alarm-action-btn" id="btn-arm-night" disabled>Arm Night</button>',
                '      <button class="alarm-action-btn" id="btn-disarm" disabled>Disarm</button>',
                '    </div>',
                '  </div>',
                '  <div class="alarm-error" id="alarm-error" style="display:none;"></div>',
                '</div>'
            ].join('\n');

            container.appendChild(viewElement);

            this._setupEventListeners();
            this._initController();
        },

        unmount: function() {
            console.log('[View] Unmounting AlarmView');

            // A5.6: Reset listener flag for clean re-mount
            this._listenersSetup = false;

            var element = document.getElementById(this.id);
            if (element) {
                element.remove();
            }
        },

        update: function(data) {
            console.log('[View] Updating AlarmView', data);

            var alarmState = data.alarmState || {};

            if (!alarmState.isHydrated) {
                this._renderLoading();
                return;
            }

            this._clearError();
            // A5.3: Clear action status when state updates (poll recovered)
            this._hideActionStatus();
            // A5.3: Update connection warning based on poll failures
            this._updateConnectionWarning();
            this._renderModeUI(alarmState);
            this._renderDelay(alarmState.delay);
            this._renderMeta(alarmState);
            this._updateButtonStates(alarmState);
        },

        // ====================================================================
        // Private Methods
        // ====================================================================

        _setupEventListeners: function() {
            var self = this;

            // A5.6: Prevent duplicate listeners
            if (this._listenersSetup) {
                console.log('[AlarmView] Listeners already setup, skipping');
                return;
            }

            var btnArmHome = document.getElementById('btn-arm-home');
            var btnArmAway = document.getElementById('btn-arm-away');
            var btnArmNight = document.getElementById('btn-arm-night');
            var btnDisarm = document.getElementById('btn-disarm');

            if (btnArmHome) {
                btnArmHome.addEventListener('click', function() {
                    self._handleAction('arm_home');
                });
            }

            if (btnArmAway) {
                btnArmAway.addEventListener('click', function() {
                    self._handleAction('arm_away');
                });
            }

            if (btnArmNight) {
                btnArmNight.addEventListener('click', function() {
                    self._handleAction('arm_night');
                });
            }

            if (btnDisarm) {
                btnDisarm.addEventListener('click', function() {
                    self._handleAction('disarm');
                });
            }

            this._listenersSetup = true;
        },

        _initController: function() {
            if (!window.SmartDisplay.alarmController) {
                console.error('[AlarmView] Alarm controller not initialized');
                return;
            }

            if (!window.SmartDisplay.alarmController.currentState) {
                window.SmartDisplay.alarmController.init()
                    .catch(function(err) {
                        console.error('[AlarmView] Failed to init controller:', err);
                    });
            }
        },

        _handleAction: function(action) {
            var self = this;
            var controller = window.SmartDisplay.alarmController;
            var strings = window.SmartDisplay.strings;

            if (!controller) {
                this._showError(strings.error.controllerNotInit);
                return;
            }

            // A5.2: Prevent action if already loading (spam protection)
            if (controller.isLoading) {
                console.log('[AlarmView] Action ignored: already processing');
                return;
            }

            console.log('[AlarmView] Action requested:', action);

            // Disable all buttons immediately
            this._setButtonsDisabled(true);
            this._showActionStatus(strings.alarm.action.sending);
            this._clearError();

            controller.requestAction(action)
                .then(function(response) {
                    console.log('[AlarmView] Action accepted:', response);
                    self._showActionStatus(strings.alarm.action.waiting);
                })
                .catch(function(err) {
                    console.error('[AlarmView] Action failed:', err);
                    
                    // Handle specific error codes using strings
                    if (err.statusCode === 409) {
                        self._showError(strings.alarm.action.blocked);
                    } else if (err.statusCode === 503) {
                        self._showError(strings.alarm.action.unreachable);
                    } else if (err.statusCode === 400) {
                        self._showError(strings.alarm.action.invalid);
                    } else {
                        self._showError(err.message || strings.alarm.action.failed);
                    }

                    self._updateButtonStates(window.SmartDisplay.store.getState().alarmState);
                    self._hideActionStatus();
                });
        },

        _setButtonsDisabled: function(disabled) {
            var buttons = [
                document.getElementById('btn-arm-home'),
                document.getElementById('btn-arm-away'),
                document.getElementById('btn-arm-night'),
                document.getElementById('btn-disarm')
            ];

            buttons.forEach(function(btn) {
                if (btn) {
                    btn.disabled = disabled;
                }
            });
        },

        _updateButtonStates: function(alarmState) {
            if (!alarmState || !alarmState.isHydrated) {
                this._setButtonsDisabled(true);
                return;
            }

            var state = (alarmState.state || 'unknown').toLowerCase();
            var triggered = alarmState.triggered;

            var btnArmHome = document.getElementById('btn-arm-home');
            var btnArmAway = document.getElementById('btn-arm-away');
            var btnArmNight = document.getElementById('btn-arm-night');
            var btnDisarm = document.getElementById('btn-disarm');
            var actionsContainer = document.getElementById('alarm-actions');

            // Hide actions entirely if triggered
            if (triggered || state === 'triggered') {
                if (actionsContainer) {
                    actionsContainer.style.display = 'none';
                }
                return;
            }

            // Show actions container
            if (actionsContainer) {
                actionsContainer.style.display = 'block';
            }

            // Disable all during pending/arming states
            if (state === 'pending' || state === 'arming') {
                this._setButtonsDisabled(true);
                return;
            }

            // Enable arm buttons only when disarmed
            var isDisarmed = state === 'disarmed';
            if (btnArmHome) btnArmHome.disabled = !isDisarmed;
            if (btnArmAway) btnArmAway.disabled = !isDisarmed;
            if (btnArmNight) btnArmNight.disabled = !isDisarmed;

            // Enable disarm button only when armed
            var isArmed = state.indexOf('armed_') === 0;
            if (btnDisarm) btnDisarm.disabled = !isArmed;
        },

        _showActionStatus: function(message) {
            var statusEl = document.getElementById('action-status');
            if (statusEl) {
                statusEl.textContent = message;
                statusEl.style.display = 'block';
            }
        },

        _hideActionStatus: function() {
            var statusEl = document.getElementById('action-status');
            if (statusEl) {
                statusEl.style.display = 'none';
                statusEl.textContent = '';
            }
        },

        _updateConnectionWarning: function() {
            var warningEl = document.getElementById('alarm-connection-warning');
            if (!warningEl) {
                return;
            }

            var controller = window.SmartDisplay.alarmController;
            if (controller && controller.pollFailureCount >= 3) {
                var strings = window.SmartDisplay.strings;
                warningEl.textContent = strings.connection.retrying;
                warningEl.style.display = 'block';
            } else {
                warningEl.style.display = 'none';
            }
        },

        _renderLoading: function() {
            var modeEl = document.getElementById('alarm-mode');
            var messageEl = document.getElementById('alarm-message');
            var contextEl = document.getElementById('alarm-context');
            var countdownEl = document.getElementById('alarm-countdown-container');

            if (modeEl) {
                modeEl.textContent = 'Connecting...';
                modeEl.className = 'alarm-mode mode-unknown';
            }

            if (messageEl) {
                messageEl.textContent = 'Waiting for Alarmo state...';
            }

            if (contextEl) {
                contextEl.textContent = 'Fetching latest data';
            }

            if (countdownEl) {
                countdownEl.style.display = 'none';
            }
        },

        _renderModeUI: function(alarmState) {
            var stateKey = (alarmState.state || 'unknown').toLowerCase();
            var modeEl = document.getElementById('alarm-mode');
            var messageEl = document.getElementById('alarm-message');

            if (modeEl) {
                modeEl.textContent = this._formatModeName(stateKey, alarmState.triggered);
                modeEl.className = 'alarm-mode mode-' + stateKey.replace(/[^a-z0-9_]/g, '-');
            }

            if (messageEl) {
                messageEl.textContent = this._renderMessage(alarmState);
            }
        },

        _renderDelay: function(delay) {
            var containerEl = document.getElementById('alarm-countdown-container');
            var labelEl = document.getElementById('countdown-label');
            var valueEl = document.getElementById('countdown-value');

            if (!containerEl) {
                return;
            }

            if (!delay || typeof delay.remaining !== 'number') {
                containerEl.style.display = 'none';
                return;
            }

            containerEl.style.display = 'flex';

            if (labelEl) {
                labelEl.textContent = this._formatDelayLabel(delay);
            }

            if (valueEl) {
                valueEl.textContent = this._formatDelayValue(delay);
            }
        },

        _renderMeta: function(alarmState) {
            var contextEl = document.getElementById('alarm-context');

            if (!contextEl) {
                return;
            }

            var explanation = this._getStateExplanation(alarmState);
            
            if (explanation) {
                contextEl.textContent = explanation;
            } else if (alarmState.triggered || alarmState.state === 'triggered') {
                contextEl.textContent = 'Triggered at ' + this._formatLastUpdated(alarmState.lastUpdated);
            } else {
                contextEl.textContent = 'Last updated ' + this._formatLastUpdated(alarmState.lastUpdated);
            }
        },

        _renderMessage: function(alarmState) {
            var state = (alarmState.state || 'unknown').toLowerCase();

            if (alarmState.triggered || state === 'triggered') {
                return 'ALARM LOCKDOWN IN EFFECT';
            }

            if (state === 'arming') {
                return 'Arming in progress';
            }

            if (state === 'pending') {
                return 'Entry/exit pending';
            }

            if (state.startsWith('armed_')) {
                var mode = state.split('_')[1];
                return 'System armed (' + mode + ')';
            }

            if (state === 'disarmed') {
                return 'System disarmed';
            }

            return 'System state: ' + state;
        },

        _getStateExplanation: function(alarmState) {
            var state = (alarmState.state || 'unknown').toLowerCase();

            if (alarmState.triggered || state === 'triggered') {
                return 'Alarm triggered.';
            }

            if (state === 'arming') {
                return 'Exit delay active.';
            }

            if (state === 'pending') {
                return 'Entry delay active.';
            }

            if (state.startsWith('armed_')) {
                return 'Alarm is armed. Disarm required to continue.';
            }

            if (state === 'disarmed') {
                return '';
            }

            return '';
        },

        _formatModeName: function(mode, triggered) {
            var modeNames = {
                'disarmed': 'Disarmed',
                'arming': 'Arming...',
                'pending': 'Pending',
                'triggered': 'ALARM TRIGGERED'
            };

            if (triggered) {
                return 'ALARM TRIGGERED';
            }

            if (modeNames[mode]) {
                return modeNames[mode];
            }

            if (mode.indexOf('armed_') === 0) {
                var suffix = mode.split('_')[1];
                return 'Armed (' + suffix.charAt(0).toUpperCase() + suffix.slice(1) + ')';
            }

            return mode.charAt(0).toUpperCase() + mode.slice(1);
        },

        _formatDelayLabel: function(delay) {
            if (!delay || !delay.type) {
                return 'Delay remaining';
            }

            return delay.type === 'entry' ? 'Entry delay' : 'Exit delay';
        },

        _formatDelayValue: function(delay) {
            if (!delay || typeof delay.remaining !== 'number') {
                return '--';
            }

            var seconds = Math.max(0, Math.floor(delay.remaining));

            if (seconds > 60) {
                return Math.ceil(seconds / 60) + 'm';
            }

            return seconds + 's';
        },

        _formatLastUpdated: function(timestamp) {
            if (!timestamp) {
                return '--';
            }

            var date = new Date(timestamp);
            if (isNaN(date.getTime())) {
                return '--';
            }

            var hours = String(date.getHours()).padStart(2, '0');
            var minutes = String(date.getMinutes()).padStart(2, '0');
            var seconds = String(date.getSeconds()).padStart(2, '0');

            return hours + ':' + minutes + ':' + seconds;
        },

        _showError: function(message) {
            var errorEl = document.getElementById('alarm-error');
            if (errorEl) {
                errorEl.textContent = 'Error: ' + message;
                errorEl.style.display = 'block';
            }
        },

        _clearError: function() {
            var errorEl = document.getElementById('alarm-error');
            if (errorEl) {
                errorEl.style.display = 'none';
                errorEl.textContent = '';
            }
        }
    };

    /**
     * Guest View
     * Guest access request and approval flow - backend-driven
     */
    var GuestView = {
        id: 'guest',
        name: 'Guest',
        countdownInterval: null,
        redirectTimeout: null,

        mount: function() {
            console.log('[View] Mounting GuestView');
            var container = document.getElementById('app');
            
            container.innerHTML = '';
            var viewElement = document.createElement('div');
            viewElement.id = this.id;
            viewElement.className = 'view view-guest';
            
            // Container with all possible states
            viewElement.innerHTML = [
                '<div class="guest-container">',
                '  <div class="guest-idle-screen" id="guest-idle-screen" style="display:none;">',
                '    <div class="guest-explanation" id="guest-explanation">--</div>',
                '    <button class="guest-action-btn guest-btn-request" id="guest-btn-request" disabled>Request Access</button>',
                '  </div>',
                '  <div class="guest-requesting-screen" id="guest-requesting-screen" style="display:none;">',
                '    <div class="guest-message" id="guest-requesting-message">Sending request...</div>',
                '    <div class="guest-countdown" id="guest-countdown-container">',
                '      <div class="countdown-label" id="countdown-label-req">Waiting for approval</div>',
                '      <div class="countdown-value" id="countdown-value-req">--</div>',
                '    </div>',
                '  </div>',
                '  <div class="guest-approved-screen" id="guest-approved-screen" style="display:none;">',
                '    <div class="guest-message" id="guest-approved-message">Access Granted</div>',
                '    <div class="guest-remaining" id="guest-remaining">--</div>',
                '    <button class="guest-action-btn guest-btn-exit" id="guest-btn-exit" disabled>Exit</button>',
                '  </div>',
                '  <div class="guest-denied-screen" id="guest-denied-screen" style="display:none;">',
                '    <div class="guest-message" id="guest-denied-message">Access Denied</div>',
                '    <div class="guest-reason" id="guest-reason">--</div>',
                '  </div>',
                '  <div class="guest-expired-screen" id="guest-expired-screen" style="display:none;">',
                '    <div class="guest-message" id="guest-expired-message">Session Expired</div>',
                '    <div class="guest-reason" id="guest-expired-reason">--</div>',
                '    <button class="guest-action-btn guest-btn-request-retry" id="guest-btn-request-retry" disabled>Try Again</button>',
                '  </div>',
                '  <div class="guest-exit-screen" id="guest-exit-screen" style="display:none;">',
                '    <div class="guest-message" id="guest-exit-message">Thank you</div>',
                '    <div class="guest-reason" id="guest-exit-reason">--</div>',
                '  </div>',
                '  <div class="guest-error" id="guest-error" style="display:none;"></div>',
                '</div>'
            ].join('\n');
            
            container.appendChild(viewElement);
            
            // Setup event listeners
            this._setupEventListeners();
            
            // Initialize controller
            this._initController();
        },

        unmount: function() {
            console.log('[View] Unmounting GuestView');
            
            // Cleanup timers
            if (this.countdownInterval) {
                clearInterval(this.countdownInterval);
                this.countdownInterval = null;
            }
            
            if (this.redirectTimeout) {
                clearTimeout(this.redirectTimeout);
                this.redirectTimeout = null;
            }
            
            var element = document.getElementById(this.id);
            if (element) {
                element.remove();
            }
        },

        update: function(data) {
            console.log('[View] Updating GuestView', data);
            
            var guestState = data.guestState || {};
            
            if (!guestState.state) {
                return; // No state data yet
            }

            // Render state-specific UI
            this._renderState(guestState);
        },

        // ====================================================================
        // Private Methods
        // ====================================================================

        _setupEventListeners: function() {
            var self = this;
            var container = document.getElementById(this.id);

            if (!container) return;

            // Request access button
            var requestBtn = document.getElementById('guest-btn-request');
            if (requestBtn) {
                requestBtn.addEventListener('click', function() {
                    self._handleRequestAccess();
                });
            }

            // Exit button
            var exitBtn = document.getElementById('guest-btn-exit');
            if (exitBtn) {
                exitBtn.addEventListener('click', function() {
                    self._handleExit();
                });
            }

            // Retry button
            var retryBtn = document.getElementById('guest-btn-request-retry');
            if (retryBtn) {
                retryBtn.addEventListener('click', function() {
                    self._handleRequestAccess();
                });
            }
        },

        _initController: function() {
            var self = this;

            if (!window.SmartDisplay.guestController) {
                console.error('[GuestView] Guest controller not initialized');
                return;
            }

            // Initialize if not done
            if (!window.SmartDisplay.guestController.currentState) {
                window.SmartDisplay.guestController.init()
                    .catch(function(err) {
                        console.error('[GuestView] Failed to init controller:', err);
                    });
            }
        },

        _renderState: function(guestState) {
            var state = guestState.state || 'unknown';
            
            console.log('[GuestView] Rendering state:', state);

            // Hide all screens
            this._hideAllScreens();

            // Cleanup countdown
            if (this.countdownInterval) {
                clearInterval(this.countdownInterval);
                this.countdownInterval = null;
            }

            // Cleanup redirect timeout
            if (this.redirectTimeout) {
                clearTimeout(this.redirectTimeout);
                this.redirectTimeout = null;
            }

            this._clearError();

            // Render appropriate state
            switch (state) {
                case 'GuestIdle':
                    this._renderIdle(guestState);
                    break;
                case 'GuestRequesting':
                    this._renderRequesting(guestState);
                    break;
                case 'GuestApproved':
                    this._renderApproved(guestState);
                    break;
                case 'GuestDenied':
                    this._renderDenied(guestState);
                    break;
                case 'GuestExpired':
                    this._renderExpired(guestState);
                    break;
                case 'GuestExit':
                    this._renderExit(guestState);
                    break;
                default:
                    console.warn('[GuestView] Unknown state:', state);
            }
        },

        _renderIdle: function(guestState) {
            var screenEl = document.getElementById('guest-idle-screen');
            var explanationEl = document.getElementById('guest-explanation');
            var requestBtn = document.getElementById('guest-btn-request');

            if (screenEl) screenEl.style.display = 'flex';

            if (explanationEl && guestState.explanation) {
                explanationEl.textContent = guestState.explanation;
            }

            if (requestBtn) {
                requestBtn.disabled = false;
            }
        },

        _renderRequesting: function(guestState) {
            var self = this;
            var screenEl = document.getElementById('guest-requesting-screen');
            var messageEl = document.getElementById('guest-requesting-message');

            if (screenEl) screenEl.style.display = 'flex';

            if (messageEl && guestState.message) {
                messageEl.textContent = guestState.message;
            }

            // Show countdown if available
            if (guestState.timeoutSeconds !== undefined && guestState.timeoutSeconds > 0) {
                this._renderCountdown(guestState.timeoutSeconds, 'Waiting for approval');
            }
        },

        _renderApproved: function(guestState) {
            var screenEl = document.getElementById('guest-approved-screen');
            var messageEl = document.getElementById('guest-approved-message');
            var remainingEl = document.getElementById('guest-remaining');
            var exitBtn = document.getElementById('guest-btn-exit');

            if (screenEl) screenEl.style.display = 'flex';

            if (messageEl && guestState.message) {
                messageEl.textContent = guestState.message;
            }

            // Show remaining time if available
            if (remainingEl && guestState.remainingSeconds !== undefined) {
                var remaining = guestState.remainingSeconds;
                if (remaining > 60) {
                    var minutes = Math.ceil(remaining / 60);
                    remainingEl.textContent = 'Access expires in ' + minutes + ' minute' + (minutes !== 1 ? 's' : '');
                } else {
                    remainingEl.textContent = 'Access expires in ' + remaining + ' second' + (remaining !== 1 ? 's' : '');
                }
            }

            if (exitBtn) {
                exitBtn.disabled = false;
            }
        },

        _renderDenied: function(guestState) {
            var screenEl = document.getElementById('guest-denied-screen');
            var messageEl = document.getElementById('guest-denied-message');
            var reasonEl = document.getElementById('guest-reason');

            if (screenEl) screenEl.style.display = 'flex';

            if (messageEl && guestState.message) {
                messageEl.textContent = guestState.message;
            }

            if (reasonEl && guestState.reason) {
                reasonEl.textContent = guestState.reason;
            } else if (reasonEl) {
                reasonEl.textContent = 'Access was denied by the owner.';
            }
        },

        _renderExpired: function(guestState) {
            var screenEl = document.getElementById('guest-expired-screen');
            var messageEl = document.getElementById('guest-expired-message');
            var reasonEl = document.getElementById('guest-expired-reason');
            var retryBtn = document.getElementById('guest-btn-request-retry');

            if (screenEl) screenEl.style.display = 'flex';

            if (messageEl && guestState.message) {
                messageEl.textContent = guestState.message;
            }

            if (reasonEl && guestState.reason) {
                reasonEl.textContent = guestState.reason;
            } else if (reasonEl) {
                reasonEl.textContent = 'Your session has expired. You can request access again.';
            }

            if (retryBtn) {
                retryBtn.disabled = false;
            }
        },

        _renderExit: function(guestState) {
            var self = this;
            var screenEl = document.getElementById('guest-exit-screen');
            var messageEl = document.getElementById('guest-exit-message');
            var reasonEl = document.getElementById('guest-exit-reason');

            if (screenEl) screenEl.style.display = 'flex';

            if (messageEl && guestState.message) {
                messageEl.textContent = guestState.message;
            }

            if (reasonEl && guestState.reason) {
                reasonEl.textContent = guestState.reason;
            } else if (reasonEl) {
                reasonEl.textContent = 'You have successfully exited guest mode.';
            }

            // Auto-redirect to home after delay
            this.redirectTimeout = setTimeout(function() {
                console.log('[GuestView] Auto-redirecting to home');
                window.SmartDisplay.store.setState({
                    guestState: {
                        isGuestActive: false
                    },
                    menu: {
                        currentView: 'home'
                    }
                });
            }, 3000); // 3 second delay
        },

        _renderCountdown: function(seconds, label) {
            var containerEl = document.getElementById('guest-countdown-container');
            var labelEl = document.getElementById('countdown-label-req');
            var valueEl = document.getElementById('countdown-value-req');

            if (!containerEl) return;

            // Check for reduced motion preference
            var prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

            // Show container
            containerEl.style.display = 'flex';

            // Update label
            if (labelEl && label) {
                labelEl.textContent = label;
            }

            // Update value
            if (valueEl) {
                if (seconds > 60) {
                    var minutes = Math.ceil(seconds / 60);
                    valueEl.textContent = minutes + 'm';
                } else {
                    valueEl.textContent = seconds + 's';
                }
            }

            // For reduced motion, just update once instead of animating
            if (prefersReducedMotion) {
                if (this.countdownInterval) {
                    clearInterval(this.countdownInterval);
                    this.countdownInterval = null;
                }
                return;
            }

            // Update countdown every second for dynamic display
            if (this.countdownInterval) {
                clearInterval(this.countdownInterval);
            }

            this.countdownInterval = setInterval(function() {
                seconds--;
                if (seconds < 0) {
                    clearInterval(this.countdownInterval);
                    this.countdownInterval = null;
                    return;
                }

                if (valueEl) {
                    if (seconds > 60) {
                        var minutes = Math.ceil(seconds / 60);
                        valueEl.textContent = minutes + 'm';
                    } else {
                        valueEl.textContent = seconds + 's';
                    }
                }
            }.bind(this), 1000);
        },

        _hideAllScreens: function() {
            var screens = [
                'guest-idle-screen',
                'guest-requesting-screen',
                'guest-approved-screen',
                'guest-denied-screen',
                'guest-expired-screen',
                'guest-exit-screen'
            ];

            screens.forEach(function(screenId) {
                var el = document.getElementById(screenId);
                if (el) el.style.display = 'none';
            });
        },

        _handleRequestAccess: function() {
            var self = this;
            var controller = window.SmartDisplay.guestController;

            if (!controller) {
                console.error('[GuestView] Controller not available');
                return;
            }

            console.log('[GuestView] Requesting access');

            this._disableAllButtons();
            this._clearError();

            controller.requestAccess()
                .then(function() {
                    console.log('[GuestView] Request successful');
                    // View will update via state subscription
                })
                .catch(function(err) {
                    console.error('[GuestView] Request failed:', err);
                    self._showError(err.message || 'Failed to request access');
                    self._enableAllButtons();
                });
        },

        _handleExit: function() {
            var self = this;
            var controller = window.SmartDisplay.guestController;

            if (!controller) {
                console.error('[GuestView] Controller not available');
                return;
            }

            console.log('[GuestView] Exiting guest mode');

            this._disableAllButtons();
            this._clearError();

            controller.exitGuest()
                .then(function() {
                    console.log('[GuestView] Exit successful');
                    // View will update via state subscription
                })
                .catch(function(err) {
                    console.error('[GuestView] Exit failed:', err);
                    self._showError(err.message || 'Failed to exit');
                    self._enableAllButtons();
                });
        },

        _disableAllButtons: function() {
            var buttons = document.querySelectorAll('.guest-action-btn');
            buttons.forEach(function(btn) {
                btn.disabled = true;
            });
        },

        _enableAllButtons: function() {
            var buttons = document.querySelectorAll('.guest-action-btn');
            buttons.forEach(function(btn) {
                // Re-enable based on current state (check if it should be enabled)
                // For now, just enable all non-disabled ones
                if (!btn.hasAttribute('data-permanently-disabled')) {
                    btn.disabled = false;
                }
            });
        },

        _showError: function(message) {
            var errorEl = document.getElementById('guest-error');
            if (errorEl) {
                errorEl.textContent = 'Error: ' + message;
                errorEl.style.display = 'block';
            }
        },

        _clearError: function() {
            var errorEl = document.getElementById('guest-error');
            if (errorEl) {
                errorEl.style.display = 'none';
                errorEl.textContent = '';
            }
        }
    };

    /**
     * Settings View
     * System settings and diagnostics (admin only)
     */
    var SettingsView = {
        id: 'settings',
        name: 'Settings',

        mount: function() {
            console.log('[View] Mounting SettingsView');
            var container = document.getElementById('app');
            
            container.innerHTML = '';
            var viewElement = document.createElement('div');
            viewElement.id = this.id;
            viewElement.className = 'view view-settings';
            
            viewElement.innerHTML = [
                '<div class="view-content">',
                '  <div class="settings-section">',
                '    <h2>System Status</h2>',
                '    <div class="settings-item">',
                '      <span class="settings-label">Backend:</span>',
                '      <span class="settings-value" id="settings-backend">--</span>',
                '    </div>',
                '    <div class="settings-item">',
                '      <span class="settings-label">Home Assistant:</span>',
                '      <span class="settings-value" id="settings-ha">--</span>',
                '    </div>',
                '    <div class="settings-item">',
                '      <span class="settings-label">Platform:</span>',
                '      <span class="settings-value" id="settings-platform">--</span>',
                '    </div>',
                '  </div>',
                '</div>',
                '<div class="view-nav">',
                '  <button class="nav-btn" data-view="home">Home</button>',
                '  <button class="nav-btn" data-view="alarm">Alarm</button>',
                '  <button class="nav-btn" data-view="devices">Devices</button>',
                '</div>'
            ].join('\n');
            
            container.appendChild(viewElement);
        },

        unmount: function() {
            console.log('[View] Unmounting SettingsView');
            var element = document.getElementById(this.id);
            if (element) {
                element.remove();
            }
        },

        update: function(data) {
            console.log('[View] Updating SettingsView', data);
            // Settings view updates would populate status information
        }
    };

    /**
     * Menu View
     * Navigation menu - backend-driven with role/state awareness
     */
    var MenuView = {
        id: 'menu',
        name: 'Menu',
        currentView: null,
        isInitialized: false,

        mount: function() {
            console.log('[View] Mounting MenuView as overlay');
            var overlay = document.getElementById('menu-overlay');
            
            if (!overlay) {
                console.error('[MenuView] Overlay container not found');
                return;
            }

            // Clear overlay content
            overlay.innerHTML = '';
            
            var viewElement = document.createElement('div');
            viewElement.id = this.id;
            viewElement.className = 'view view-menu';
            
            viewElement.innerHTML = [
                '<div class="menu-container">',
                '  <div class="menu-header">',
                '    <h2 class="menu-title">Menu</h2>',
                '  </div>',
                '  <div class="menu-content" id="menu-content">',
                '    <!-- Menu sections rendered here -->',
                '  </div>',
                '  <div class="menu-error" id="menu-error" style="display:none;"></div>',
                '  <div class="menu-footer">',
                '    <button class="menu-close-btn" id="menu-close-btn">Close</button>',
                '  </div>',
                '</div>'
            ].join('\n');
            
            overlay.appendChild(viewElement);
            
            // Setup event listeners
            this._setupEventListeners();
            
            // Initialize controller and render menu
            this._initController();
            
            this.isInitialized = true;
        },

        unmount: function() {
            console.log('[View] Unmounting MenuView (hiding overlay)');
            var overlay = document.getElementById('menu-overlay');
            if (overlay) {
                overlay.style.display = 'none';
            }
        },

        update: function(data) {
            console.log('[View] Updating MenuView', data);
            
            // Update current view highlight
            if (data.menu && data.menu.currentView) {
                this.currentView = data.menu.currentView;
                this._updateHighlight();
            }
        },

        // ====================================================================
        // Private Methods
        // ====================================================================

        _setupEventListeners: function() {
            var self = this;
            var container = document.getElementById(this.id);

            if (!container) return;

            // Close button
            var closeBtn = document.getElementById('menu-close-btn');
            if (closeBtn) {
                closeBtn.addEventListener('click', function() {
                    self._handleClose();
                });
            }

            // Click outside menu to close (on overlay background)
            var overlay = document.getElementById('menu-overlay');
            if (overlay) {
                overlay.addEventListener('click', function(e) {
                    // Only close if clicking on overlay background, not the menu itself
                    if (e.target === overlay) {
                        self._handleClose();
                    }
                });
            }

            // Menu item clicks
            container.addEventListener('click', function(e) {
                var menuItem = e.target.closest('[data-view]');
                if (menuItem && !menuItem.disabled) {
                    self._handleMenuItemClick(menuItem);
                }
            });
        },

        _initController: function() {
            var self = this;

            if (!window.SmartDisplay.menuController) {
                console.error('[MenuView] Menu controller not initialized');
                return;
            }

            // Initialize if not done
            if (!window.SmartDisplay.menuController.menuData) {
                window.SmartDisplay.menuController.init()
                    .then(function() {
                        self._renderMenu();
                    })
                    .catch(function(err) {
                        console.error('[MenuView] Failed to init controller:', err);
                        self._showError('Failed to load menu');
                        self._renderFallback();
                    });
            } else {
                this._renderMenu();
            }
        },

        _renderMenu: function() {
            var controller = window.SmartDisplay.menuController;
            var sections = controller.getVisibleSections();

            if (!sections || sections.length === 0) {
                console.warn('[MenuView] No visible sections');
                this._renderFallback();
                return;
            }

            var contentEl = document.getElementById('menu-content');
            if (!contentEl) return;

            contentEl.innerHTML = '';

            // Render each section
            sections.forEach(function(section) {
                var sectionEl = document.createElement('div');
                sectionEl.className = 'menu-section';

                // Section header
                if (section.label) {
                    var headerEl = document.createElement('div');
                    headerEl.className = 'menu-section-header';
                    headerEl.textContent = section.label;
                    sectionEl.appendChild(headerEl);
                }

                // Section items
                if (section.items && Array.isArray(section.items)) {
                    var itemsContainerEl = document.createElement('div');
                    itemsContainerEl.className = 'menu-items';

                    section.items.forEach(function(item) {
                        // Skip disabled items - don't render them
                        if (!controller.isItemEnabled(item)) {
                            return;
                        }

                        var itemEl = document.createElement('button');
                        itemEl.className = 'menu-item';
                        itemEl.setAttribute('data-view', item.view);

                        // Add item label
                        var labelEl = document.createElement('span');
                        labelEl.className = 'menu-item-label';
                        labelEl.textContent = item.label || item.view;
                        itemEl.appendChild(labelEl);

                        // Add badge/state if provided
                        if (item.badge) {
                            var badgeEl = document.createElement('span');
                            badgeEl.className = 'menu-item-badge';
                            badgeEl.textContent = item.badge;
                            itemEl.appendChild(badgeEl);
                        }

                        // Add state class if provided
                        if (item.state) {
                            itemEl.className += ' menu-item-' + item.state;
                        }

                        itemsContainerEl.appendChild(itemEl);
                    });

                    sectionEl.appendChild(itemsContainerEl);
                }

                contentEl.appendChild(sectionEl);
            });

            // Update highlights
            this._updateHighlight();
        },

        _renderFallback: function() {
            var contentEl = document.getElementById('menu-content');
            if (!contentEl) return;

            contentEl.innerHTML = [
                '<div class="menu-section">',
                '  <div class="menu-items">',
                '    <button class="menu-item" data-view="home">Home</button>',
                '    <button class="menu-item" data-view="alarm">Alarm</button>',
                '  </div>',
                '</div>'
            ].join('\n');

            // Re-setup event listeners for fallback items
            var self = this;
            contentEl.addEventListener('click', function(e) {
                var menuItem = e.target.closest('[data-view]');
                if (menuItem) {
                    self._handleMenuItemClick(menuItem);
                }
            });
        },

        _updateHighlight: function() {
            var currentView = this.currentView;
            var menuItems = document.querySelectorAll('[data-view]');

            menuItems.forEach(function(item) {
                var view = item.getAttribute('data-view');
                if (view === currentView) {
                    item.className += ' menu-item-active';
                } else {
                    item.className = item.className.replace(' menu-item-active', '');
                }
            });
        },

        _handleMenuItemClick: function(btn) {
            var viewId = btn.getAttribute('data-view');

            if (!viewId) return;

            if (window.SmartDisplay.viewManager && window.SmartDisplay.viewManager.isAlarmLocked()) {
                console.log('[MenuView] Alarm lock active, ignoring menu navigation');
                return;
            }

            console.log('[MenuView] Menu item clicked:', viewId);

            // Update store and trigger navigation
            if (window.SmartDisplay.store) {
                window.SmartDisplay.store.setState({
                    menu: {
                        currentView: viewId
                    }
                });
            }
        },

        _handleClose: function() {
            console.log('[MenuView] Close button clicked');
            // Close the menu overlay
            if (window.SmartDisplay.viewManager) {
                window.SmartDisplay.viewManager.closeMenu();
            }
        },

        _showError: function(message) {
            var errorEl = document.getElementById('menu-error');
            if (errorEl) {
                errorEl.textContent = message;
                errorEl.style.display = 'block';
            }
        }
    };

    // ========================================================================
    // View Manager
    // ========================================================================
    var ViewManager = {
        currentView: null,
        views: {
            'first-boot': FirstBootView,
            'home': HomeView,
            'alarm': AlarmView,
            'guest': GuestView,
            'settings': SettingsView,
            'menu': MenuView
        },

        // ====================================================================
        // View Switching
        // ====================================================================

        /**
         * Switch to a specific view
         * @param {string} viewId - View identifier
         * @param {object} data - Data to pass to view update
         */
        switchToView: function(viewId, data) {
            var view = this.views[viewId];
            
            if (!view) {
                console.error('[ViewManager] View not found: ' + viewId);
                return false;
            }

            if (this.currentView && this.currentView.id !== viewId) {
                console.log('[ViewManager] Switching from ' + this.currentView.id + ' to ' + viewId);
                this.currentView.unmount();
            } else if (this.currentView && this.currentView.id === viewId) {
                // Same view, just update
                if (data) {
                    this.currentView.update(data);
                }
                return true;
            }

            this.currentView = view;
            view.mount();
            
            if (data) {
                view.update(data);
            }

            return true;
        },

        /**
         * Update current view with new data
         * @param {object} data - Data to update view with
         */
        updateCurrentView: function(data) {
            if (this.currentView) {
                this.currentView.update(data);
            }
        },

        // ====================================================================
        // Routing Logic
        // ====================================================================

        /**
         * Determine which view should be displayed based on state
         * Routing priority:
         * 1. FirstBoot → FirstBootView
         * 2. Menu requested → MenuView
         * 3. Guest active and not admin → GuestView
         * 4. Alarm in alert or critical state → AlarmView
         * 5. Menu.currentView = 'settings' and admin → SettingsView
         * 6. Otherwise → HomeView (or determined by menu.currentView)
         * 
         * @param {object} state - Full application state
         * @returns {string} - View ID to display
         */
        getNextView: function(state) {
            if (state.firstBoot) {
                console.log('[ViewManager] Route: FirstBoot');
                return 'first-boot';
            }

            var alarmState = state.alarmState || {};

            if (!alarmState.isHydrated || this._shouldLockToAlarm(alarmState)) {
                console.log('[ViewManager] Route: Alarm (locked state)');
                return 'alarm';
            }

            if (state.menu && state.menu.currentView === 'menu') {
                console.log('[ViewManager] Route: Menu');
                return 'menu';
            }

            if (state.guestState && state.guestState.isGuestActive) {
                console.log('[ViewManager] Route: Guest');
                return 'guest';
            }

            if (state.menu && state.menu.currentView === 'settings') {
                console.log('[ViewManager] Route: Settings');
                return 'settings';
            }

            if (state.menu && state.menu.currentView === 'alarm') {
                console.log('[ViewManager] Route: Alarm (user view)');
                return 'alarm';
            }

            console.log('[ViewManager] Route: Home (default)');
            return 'home';
        },

        /**
         * Render view based on current state
         * Compares next view with current and switches if needed
         */
        render: function() {
            if (!window.SmartDisplay.store) {
                console.warn('[ViewManager] Store not initialized yet');
                return;
            }

            var state = window.SmartDisplay.store.getState();
            var nextViewId = this.getNextView(state);

            this._applyAlarmLock(state.alarmState || {});

            // Switch view if different
            if (!this.currentView || this.currentView.id !== nextViewId) {
                this.switchToView(nextViewId, state);
            } else {
                // Same view, update with new state
                this.updateCurrentView(state);
            }
        },

        _shouldLockToAlarm: function(alarmState) {
            if (!alarmState || !alarmState.state) {
                return false;
            }

            var normalized = alarmState.state.toLowerCase();

            if (alarmState.triggered || normalized === 'triggered') {
                return true;
            }

            if (normalized === 'arming' || normalized === 'pending') {
                return true;
            }

            if (normalized.startsWith('armed_')) {
                return true;
            }

            return false;
        },

        _applyAlarmLock: function(alarmState) {
            var locked = !alarmState || !alarmState.isHydrated || this._shouldLockToAlarm(alarmState);
            var overlay = document.getElementById('menu-overlay');

            if (document && document.body) {
                document.body.classList.toggle('alarm-locked', locked);
            }

            if (overlay) {
                overlay.classList.toggle('menu-locked', locked);
                if (locked) {
                    this.closeMenu();
                }
            }
        },

        isAlarmLocked: function() {
            if (!window.SmartDisplay || !window.SmartDisplay.store) {
                return false;
            }

            var alarmState = window.SmartDisplay.store.getState().alarmState;
            return !alarmState || !alarmState.isHydrated || this._shouldLockToAlarm(alarmState);
        },

        // ====================================================================
        // Initialization
        // ====================================================================

        /**
         * Initialize view manager and subscribe to state changes
         */
        init: function() {
            var self = this;

            console.log('[ViewManager] Initializing');

            // Mount MenuView as global overlay (only once)
            MenuView.mount();

            // Subscribe to state changes
            if (window.SmartDisplay.store) {
                window.SmartDisplay.store.subscribe(function(updates) {
                    console.log('[ViewManager] State changed, re-rendering');
                    self.render();
                    
                    // Update menu visibility based on state
                    if (updates.menu && updates.menu.isOpen !== undefined) {
                        if (updates.menu.isOpen) {
                            self.openMenu();
                        } else {
                            self.closeMenu();
                        }
                    }
                });
            }

            // Initial render
            this.render();

            // Listen for navigation button clicks
            document.addEventListener('click', function(e) {
                var navBtn = e.target.closest('[data-view]');
                if (navBtn) {
                    if (self.isAlarmLocked()) {
                        console.log('[ViewManager] Navigation blocked while alarm lock is active');
                        return;
                    }

                    var viewId = navBtn.getAttribute('data-view');
                    console.log('[ViewManager] Navigation clicked: ' + viewId);
                    
                    // Update menu state
                    if (window.SmartDisplay.store) {
                        window.SmartDisplay.store.setState({
                            menu: {
                                currentView: viewId
                            }
                        });
                    }
                }
            });
        },

        // ====================================================================
        // Menu Overlay Control
        // ====================================================================

        /**
         * Show menu overlay
         */
        openMenu: function() {
            if (this.isAlarmLocked()) {
                console.log('[ViewManager] Menu locked by alarm state');
                return;
            }

            console.log('[ViewManager] Opening menu');
            var overlay = document.getElementById('menu-overlay');
            if (overlay) {
                overlay.classList.add('menu-open');
            }
        },

        /**
         * Hide menu overlay
         */
        closeMenu: function() {
            console.log('[ViewManager] Closing menu');
            var overlay = document.getElementById('menu-overlay');
            if (overlay) {
                overlay.classList.remove('menu-open');
            }
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.viewManager = ViewManager;

    // Auto-initialize when store is ready
    if (window.SmartDisplay.onReady) {
        window.SmartDisplay.onReady(function() {
            ViewManager.init();
        });
    }

    console.log('[SmartDisplay] View manager registered');

})();
