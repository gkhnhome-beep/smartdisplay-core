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
     * Login View
     * FAZ L1: PIN-based authentication
     * Fullscreen PIN entry view
     */
    var LoginView = {
        id: 'login',
        name: 'Login',

        mount: function() {
            console.log('[View] Mounting LoginView');
            var container = document.getElementById('app');
            
            container.innerHTML = '';
            var viewElement = document.createElement('div');
            viewElement.id = this.id;
            viewElement.className = 'view view-login';
            
            viewElement.innerHTML = [
                '<div class="view-content">',
                '  <div class="login-container">',
                '    <div class="login-header">',
                '      <h1>SmartDisplay</h1>',
                '      <p>Enter PIN to continue</p>',
                '    </div>',
                '    <div class="login-pin-display" id="login-pin-display">',
                '      <div class="pin-dots"></div>',
                '    </div>',
                '    <div class="login-keypad">',
                '      <button class="key-btn" data-digit="1">1</button>',
                '      <button class="key-btn" data-digit="2">2</button>',
                '      <button class="key-btn" data-digit="3">3</button>',
                '      <button class="key-btn" data-digit="4">4</button>',
                '      <button class="key-btn" data-digit="5">5</button>',
                '      <button class="key-btn" data-digit="6">6</button>',
                '      <button class="key-btn" data-digit="7">7</button>',
                '      <button class="key-btn" data-digit="8">8</button>',
                '      <button class="key-btn" data-digit="9">9</button>',
                '      <button class="key-btn key-btn-clear" id="btn-clear">Clear</button>',
                '      <button class="key-btn" data-digit="0">0</button>',
                '      <button class="key-btn key-btn-back" id="btn-back">⌫</button>',
                '    </div>',
                '    <div class="login-footer">',
                '      <p class="login-hint">Default PIN: 1234 (admin)</p>',
                '      <button id="btn-guest-access" class="btn-guest-access">Request Guest Access</button>',
                '    </div>',
                '  </div>',
                '</div>'
            ].join('\n');
            
            container.appendChild(viewElement);
            
            // Initialize controller
            if (window.SmartDisplay.loginController) {
                window.SmartDisplay.loginController.init();
            }
            
            // Setup event listeners
            this._setupEventListeners();
            
            // Initial render
            this._render();
        },

        unmount: function() {
            console.log('[View] Unmounting LoginView');
            var element = document.getElementById(this.id);
            if (element) {
                element.remove();
            }
        },

        update: function(data) {
            console.log('[View] Updating LoginView', data);
            this._render();
        },

        _setupEventListeners: function() {
            var self = this;
            var controller = window.SmartDisplay.loginController;

            // Number buttons
            var digitBtns = document.querySelectorAll('.key-btn[data-digit]');
            digitBtns.forEach(function(btn) {
                btn.addEventListener('click', function() {
                    var digit = this.getAttribute('data-digit');
                    controller.onDigitPress(digit);
                    self._render();
                });
            });

            // Clear button
            var clearBtn = document.getElementById('btn-clear');
            if (clearBtn) {
                clearBtn.addEventListener('click', function() {
                    controller.clear();
                    self._render();
                });
            }

            // Backspace button
            var backBtn = document.getElementById('btn-back');
            if (backBtn) {
                backBtn.addEventListener('click', function() {
                    controller.onBackspace();
                    self._render();
                });
            }

            // Guest Access button (FAZ L2)
            var guestBtn = document.getElementById('btn-guest-access');
            if (guestBtn) {
                guestBtn.addEventListener('click', function() {
                    console.log('[LoginView] Guest access button clicked');
                    ViewManager.routeToView('guest-request');
                });
            }
        },

        _render: function() {
            var controller = window.SmartDisplay.loginController;
            if (!controller) return;

            var pinDisplay = document.getElementById('login-pin-display');
            if (!pinDisplay) return;

            // Update PIN dots
            var maskedPIN = controller.getMaskedPIN();
            var dots = '';
            for (var i = 0; i < 4; i++) {
                if (i < maskedPIN.length) {
                    dots += '<span class="pin-dot filled">●</span>';
                } else {
                    dots += '<span class="pin-dot">○</span>';
                }
            }
            pinDisplay.querySelector('.pin-dots').innerHTML = dots;

            // Error state
            var container = document.querySelector('.login-container');
            if (controller.error) {
                container.classList.add('error');
                setTimeout(function() {
                    container.classList.remove('error');
                }, 500);
            }

            // Validating state
            if (controller.isValidating) {
                container.classList.add('validating');
            } else {
                container.classList.remove('validating');
            }
        }
    };

    /**
     * Guest Request View (FAZ L2)
     * Guest access request flow with HA user selection and approval waiting
     */
    var GuestRequestView = {
        id: 'guest-request',
        name: 'Guest Request',

        // Guest request state
        haUsers: [],
        selectedUser: null,
        activeRequest: null,
        countdownSeconds: 60,
        countdownInterval: null,

        mount: function() {
            console.log('[View] Mounting GuestRequestView');
            var container = document.getElementById('app');
            
            container.innerHTML = '';
            var viewElement = document.createElement('div');
            viewElement.id = this.id;
            viewElement.className = 'view view-guest-request';
            
            viewElement.innerHTML = [
                '<div class="view-content">',
                '  <div class="guest-request-container">',
                '    <div class="guest-request-header">',
                '      <h1>Request Guest Access</h1>',
                '      <p id="guest-request-status">Select a user to request access</p>',
                '    </div>',
                '    <div class="guest-request-content" id="guest-request-content">',
                '      <!-- Content will be populated based on state -->',
                '    </div>',
                '    <div class="guest-request-actions">',
                '      <button id="btn-guest-back" class="btn-secondary">Back</button>',
                '    </div>',
                '  </div>',
                '</div>'
            ].join('\n');
            
            container.appendChild(viewElement);
            
            // Setup event listeners
            this._setupEventListeners();
            
            // Load HA users and render
            this._loadHAUsers();
        },

        unmount: function() {
            console.log('[View] Unmounting GuestRequestView');
            if (this.countdownInterval) {
                clearInterval(this.countdownInterval);
                this.countdownInterval = null;
            }
            var element = document.getElementById(this.id);
            if (element) {
                element.remove();
            }
        },

        update: function(data) {
            console.log('[View] Updating GuestRequestView', data);
            this._render();
        },

        _setupEventListeners: function() {
            var self = this;
            
            // Back button
            var backBtn = document.getElementById('btn-guest-back');
            if (backBtn) {
                backBtn.addEventListener('click', function() {
                    console.log('[GuestRequestView] Back button clicked');
                    self.unmount();
                    ViewManager.routeToView('login');
                });
            }
        },

        _loadHAUsers: function() {
            var self = this;
            
            // For now, use hardcoded demo users
            // TODO: Fetch from HA API when available
            this.haUsers = [
                { id: 'user1', name: 'User 1' },
                { id: 'user2', name: 'User 2' },
                { id: 'admin', name: 'Admin' }
            ];
            
            this._render();
        },

        _render: function() {
            var content = document.getElementById('guest-request-content');
            if (!content) return;
            
            if (this.activeRequest) {
                // Render waiting/approval state
                this._renderWaitingState(content);
            } else {
                // Render user selection state
                this._renderUserSelection(content);
            }
        },

        _renderUserSelection: function(content) {
            var self = this;
            var html = '<div class="user-selection">';
            
            this.haUsers.forEach(function(user) {
                html += [
                    '<div class="user-option" data-user-id="' + user.id + '">',
                    '  <span class="user-name">' + user.name + '</span>',
                    '  <span class="user-arrow">→</span>',
                    '</div>'
                ].join('\n');
            });
            
            html += '</div>';
            content.innerHTML = html;
            
            // Setup user selection listeners
            var userOptions = content.querySelectorAll('.user-option');
            userOptions.forEach(function(option) {
                option.addEventListener('click', function() {
                    var userId = this.getAttribute('data-user-id');
                    var user = self.haUsers.find(function(u) { return u.id === userId; });
                    if (user) {
                        self._requestGuestAccess(user);
                    }
                });
            });
        },

        _renderWaitingState: function(content) {
            var self = this;
            var status = document.getElementById('guest-request-status');
            
            var html = [
                '<div class="approval-waiting">',
                '  <div class="waiting-spinner"></div>',
                '  <p class="waiting-text">Requesting approval from <strong>' + this.selectedUser.name + '</strong></p>',
                '  <div class="countdown-timer">',
                '    <span class="countdown-seconds">' + this.countdownSeconds + '</span>',
                '    <span class="countdown-label">seconds remaining</span>',
                '  </div>',
                '  <p class="waiting-help">Check your notifications on your phone</p>',
                '</div>'
            ].join('\n');
            
            content.innerHTML = html;
            
            if (status) {
                status.textContent = 'Waiting for approval...';
            }
            
            // Start countdown if not already running
            if (!this.countdownInterval) {
                this.countdownInterval = setInterval(function() {
                    self.countdownSeconds--;
                    var timer = content.querySelector('.countdown-seconds');
                    if (timer) {
                        timer.textContent = self.countdownSeconds;
                    }
                    
                    if (self.countdownSeconds <= 0) {
                        clearInterval(self.countdownInterval);
                        self.countdownInterval = null;
                        self._handleRequestExpired();
                    }
                }, 1000);
            }
        },

        _requestGuestAccess: function(user) {
            var self = this;
            this.selectedUser = user;
            
            // Call backend to create guest request (POST /api/guest/request)
            window.SmartDisplay.api.post(
                '/api/guest/request',
                { ha_user: user.id },
                {
                    onSuccess: function(response) {
                        console.log('[GuestRequestView] Guest request created:', response);
                        self.activeRequest = response.data;
                        self.countdownSeconds = 60;
                        self._render();
                        
                        // Poll for approval status
                        self._pollApprovalStatus();
                    },
                    onFailure: function(error) {
                        console.error('[GuestRequestView] Guest request failed:', error);
                        var status = document.getElementById('guest-request-status');
                        if (status) {
                            status.textContent = 'Error: ' + error.message;
                            status.classList.add('error');
                        }
                    }
                }
            );
        },

        _pollApprovalStatus: function() {
            var self = this;
            
            // Poll every 2 seconds for approval status
            var pollInterval = setInterval(function() {
                if (!self.activeRequest) {
                    clearInterval(pollInterval);
                    return;
                }
                
                // Call backend to check request status
                window.SmartDisplay.api.get(
                    '/api/ui/guest/request/' + self.activeRequest.request_id,
                    {
                        onSuccess: function(response) {
                            var req = response.data;
                            
                            if (req.status === 'approved') {
                                clearInterval(pollInterval);
                                if (self.countdownInterval) {
                                    clearInterval(self.countdownInterval);
                                    self.countdownInterval = null;
                                }
                                self._handleApproved(req);
                            } else if (req.status === 'rejected') {
                                clearInterval(pollInterval);
                                if (self.countdownInterval) {
                                    clearInterval(self.countdownInterval);
                                    self.countdownInterval = null;
                                }
                                self._handleRejected(req);
                            } else if (req.status === 'expired') {
                                clearInterval(pollInterval);
                                if (self.countdownInterval) {
                                    clearInterval(self.countdownInterval);
                                    self.countdownInterval = null;
                                }
                                self._handleRequestExpired();
                            }
                        }
                    }
                );
            }, 2000);
        },

        _handleApproved: function(request) {
            console.log('[GuestRequestView] Guest access approved:', request);
            
            // Update store to set guest as authenticated (both authState and guestState)
            if (window.SmartDisplay.store) {
                window.SmartDisplay.store.setState({
                    authState: {
                        authenticated: true,
                        role: 'guest',
                        pin: ''
                    },
                    guestState: {
                        active: true,
                        requestId: request.request_id,
                        targetUser: request.ha_user,
                        approvalTime: new Date().toISOString(),
                        pollingActive: false
                    }
                });
            }
            
            // Show approval confirmation
            this.activeRequest = null;
            var status = document.getElementById('guest-request-status');
            if (status) {
                status.textContent = 'Access approved!';
                status.classList.add('success');
            }
            
            // Delay navigation to show success state
            var self = this;
            setTimeout(function() {
                ViewManager.routeToView('home');
            }, 1500);
        },

        _handleRejected: function(request) {
            console.log('[GuestRequestView] Guest access rejected:', request);
            
            // Clear request and show rejection
            this.activeRequest = null;
            var status = document.getElementById('guest-request-status');
            if (status) {
                status.textContent = 'Access request denied';
                status.classList.add('error');
            }
            
            // Reset to user selection after delay
            var self = this;
            setTimeout(function() {
                self._render();
            }, 2000);
        },

        _handleRequestExpired: function() {
            console.log('[GuestRequestView] Guest request expired');
            
            // Clear request and show expiration
            this.activeRequest = null;
            var status = document.getElementById('guest-request-status');
            if (status) {
                status.textContent = 'Request expired. Please try again.';
                status.classList.add('error');
            }
            
            // Reset to user selection after delay
            var self = this;
            setTimeout(function() {
                self._render();
            }, 2000);
        }
    };

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
        inactivityDelay: 5000, // 5 seconds
        menuAutoHideTimeout: null,
        menuFullCloseTimeout: null, // 5-minute timeout to completely close menu
        savedMenuView: null, // Store the view that was open when menu opened

        mount: function() {
            console.log('[View] Mounting HomeView');
            var container = document.getElementById('app');
            
            container.innerHTML = '';
            var viewElement = document.createElement('div');
            viewElement.id = this.id;
            viewElement.className = 'view view-home';
            
            // Calm layout structure with clickable header and menu button
            viewElement.innerHTML = [
                '<div class="home-header" id="home-header">',
                '  <button class="home-menu-btn" id="home-menu-btn" title="Menu">☰</button>',
                '</div>',
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
            
            // Open menu automatically on mount (will auto-hide after 10s)
            console.log('[HomeView] Opening menu on initial load');
            window.SmartDisplay.viewManager.openMenu();
            this._scheduleMenuAutoHide();
            
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
            
            if (this.menuAutoHideTimeout) {
                clearTimeout(this.menuAutoHideTimeout);
                this.menuAutoHideTimeout = null;
            }
            
            if (this.menuFullCloseTimeout) {
                clearTimeout(this.menuFullCloseTimeout);
                this.menuFullCloseTimeout = null;
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
            
            // Update HA status
            if (data.haState) {
                var haStatusEl = document.getElementById('status-ha');
                if (haStatusEl) {
                    var haText = 'HA: ';
                    if (!data.haState.isConfigured) {
                        haText += 'Not configured';
                    } else if (!data.haState.isConnected) {
                        haText += 'Disconnected';
                    } else if (!data.haState.syncDone) {
                        haText += 'Connected (sync pending)';
                    } else {
                        // Show entity counts
                        var counts = data.haState.entityCounts || {};
                        var total = (counts.lights || 0) + (counts.sensors || 0) + 
                                   (counts.switches || 0) + (counts.others || 0);
                        haText += total + ' entities';
                    }
                    haStatusEl.textContent = haText;
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

            var self = this;

            // Menu button - toggle menu with auto-hide
            var menuBtn = document.getElementById('home-menu-btn');
            if (menuBtn) {
                menuBtn.addEventListener('click', function(e) {
                    e.stopPropagation();
                    console.log('[HomeView] Menu button clicked');
                    var overlay = document.getElementById('menu-overlay');
                    if (overlay && overlay.classList.contains('menu-open')) {
                        window.SmartDisplay.viewManager.closeMenu();
                    } else {
                        // Save current menu view before opening
                        var state = window.SmartDisplay.store.getState();
                        self.savedMenuView = (state.menu && state.menu.currentView) || 'home';
                        console.log('[HomeView] Saved menu view:', self.savedMenuView);
                        
                        window.SmartDisplay.viewManager.openMenu();
                        // Auto-hide menu after 10 seconds of inactivity
                        self._scheduleMenuAutoHide();
                    }
                });
            }

            // Tap anywhere (except menu button and menu) to activate
            container.addEventListener('click', function(e) {
                // Don't trigger if menu button or menu was clicked
                if (!e.target.closest('#home-menu-btn') && !e.target.closest('#menu-overlay')) {
                    self._handleTap();
                }
            });

            container.addEventListener('touchstart', function(e) {
                // Reset menu auto-hide timer on touch
                if (document.getElementById('menu-overlay').classList.contains('menu-open')) {
                    self._scheduleMenuAutoHide();
                }
                // Don't trigger if menu button or menu was touched
                if (!e.target.closest('#home-menu-btn') && !e.target.closest('#menu-overlay')) {
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
            this._scheduleInactivityTimeout();
        },

        _scheduleInactivityTimeout: function() {
            // Don't schedule inactivity timeout if menu is open
            var overlay = document.getElementById('menu-overlay');
            if (overlay && overlay.classList.contains('menu-open')) {
                console.log('[HomeView] Menu is open, skipping inactivity timeout schedule');
                return;
            }

            // Clear existing timeout
            if (this.inactivityTimeout) {
                clearTimeout(this.inactivityTimeout);
            }

            // Check for reduced motion preference
            var prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

            if (!prefersReducedMotion) {
                // Set inactivity timeout (only if not reduced motion)
                var controller = window.SmartDisplay.homeController;
                this.inactivityTimeout = setTimeout(function() {
                    if (controller) {
                        controller.setInactive();
                    }
                }, this.inactivityDelay);
            }
        },

        _scheduleMenuAutoHide: function() {
            var self = this;
            
            // Clear existing timeouts
            if (this.menuAutoHideTimeout) {
                clearTimeout(this.menuAutoHideTimeout);
            }
            if (this.menuFullCloseTimeout) {
                clearTimeout(this.menuFullCloseTimeout);
            }
            
            // First timeout: 10 seconds - show idle screen (but keep menu context)
            this.menuAutoHideTimeout = setTimeout(function() {
                console.log('[HomeView] No activity for 10s - switching to idle screen');
                // Switch to idle screen but keep menu open in background
                self._showIdleScreen();
            }, 10000); // 10 seconds
            
            // Second timeout: 5 minutes - return to home menu entirely
            this.menuFullCloseTimeout = setTimeout(function() {
                console.log('[HomeView] No activity for 5 minutes - closing menu and returning to home');
                window.SmartDisplay.viewManager.closeMenu();
                
                // Return to home menu
                if (window.SmartDisplay.store) {
                    window.SmartDisplay.store.setState({
                        menu: {
                            currentView: 'home',
                            isOpen: false
                        }
                    });
                }
            }, 300000); // 5 minutes (300 seconds)
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
            
            // If there's a saved menu view (other than home), restore the menu
            if (this.savedMenuView && this.savedMenuView !== 'home') {
                console.log('[HomeView] Restoring menu with saved view:', this.savedMenuView);
                window.SmartDisplay.viewManager.openMenu();
                
                // Set the menu view to the saved one
                if (window.SmartDisplay.store) {
                    window.SmartDisplay.store.setState({
                        menu: {
                            currentView: this.savedMenuView
                        }
                    });
                }
            }
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
                '  <div class="alarm-ha-setup-prompt" id="alarm-ha-setup-prompt" style="display:none;">',
                '    <div class="ha-setup-message">Home Assistant not fully configured.</div>',
                '    <button class="ha-setup-btn" id="alarm-ha-setup-btn">Go to Settings</button>',
                '  </div>',
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
            var haState = data.haState || {};
            console.log('[AlarmView] haState debug:', {
                isConfigured: haState.isConfigured,
                syncDone: haState.syncDone,
                fullState: haState
            });

            // Show HA setup prompt if HA configured but sync not done
            var setupPrompt = document.getElementById('alarm-ha-setup-prompt');
            if (setupPrompt) {
                if (haState.isConfigured && !haState.syncDone) {
                    console.log('[AlarmView] Showing HA setup prompt');
                    setupPrompt.style.display = 'block';
                } else {
                    setupPrompt.style.display = 'none';
                }
            }

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
            var btnHASetup = document.getElementById('alarm-ha-setup-btn');

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

            if (btnHASetup) {
                btnHASetup.addEventListener('click', function() {
                    console.log('[AlarmView] HA setup button clicked - navigating to settings');
                    window.SmartDisplay.viewManager.render('settings');
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
     * FAZ S0/S1: Settings access control and UI scaffold with Home Assistant integration surface
     */
    var SettingsView = {
        id: 'settings',
        name: 'Settings',
        currentSubpage: null,  // 'main', 'ha-settings'

        mount: function() {
            console.log('[View] Mounting SettingsView');
            var container = document.getElementById('app');
            
            container.innerHTML = '';
            var viewElement = document.createElement('div');
            viewElement.id = this.id;
            viewElement.className = 'view view-settings';
            
            viewElement.innerHTML = [
                '<div class="settings-header">',
                '  <h1 class="settings-title">Settings</h1>',
                '  <button class="settings-back-btn" id="settings-back-btn">← Home</button>',
                '</div>',
                '<div class="settings-container">',
                '  <!-- Main Settings Menu -->', 
                '  <div class="settings-main" id="settings-main">',
                '    <div class="settings-section">',
                '      <div class="settings-section-header">System</div>',
                '      <button class="settings-item-btn" id="settings-ha-btn">',
                '        <span class="settings-item-label">Home Assistant</span>',
                '        <span class="settings-item-arrow">›</span>',
                '      </button>',
                '      <button class="settings-item-btn" id="settings-alarmo-btn">',
                '        <span class="settings-item-label">Alarmo Monitoring</span>',
                '        <span class="settings-item-arrow">›</span>',
                '      </button>',
                '    </div>',
                '    <div class="settings-section">',
                '      <div class="settings-section-header">Security</div>',
                '      <button class="settings-item-btn" id="settings-alarm-config-btn">',
                '        <span class="settings-item-label">Alarm Settings</span>',
                '        <span class="settings-item-arrow">›</span>',
                '      </button>',
                '    </div>',
                '  </div>',
                '  <!-- HA Settings Subpage -->',
                '  <div class="settings-subpage" id="settings-ha-subpage" style="display:none;">',
                '    <div class="settings-section">',
                '      <div class="settings-form">',
                '        <div class="form-group">',
                '          <label for="ha-server-addr" class="form-label">Server Address</label>',
                '          <input type="text" id="ha-server-addr" class="form-input" placeholder="http://homeassistant.local:8123" />',
                '        </div>',
                '        <div class="form-group">',
                '          <label for="ha-token" class="form-label">Long-Lived Access Token</label>',
                '          <input type="password" id="ha-token" class="form-input" placeholder="••••••••••••••••••••" />',
                '        </div>',
                '        <div class="form-status">',
                '          <span class="status-label">Connection Status:</span>',
                '          <span class="status-value" id="ha-status">Not configured</span>',
                '        </div>',
                '        <div class="form-actions">',
                '          <button class="settings-save-btn" id="ha-save-btn">Save</button>',
                '          <button class="settings-test-btn" id="ha-test-btn">Test Connection</button>',
                '          <button class="settings-sync-btn" id="ha-sync-btn">Initial Sync</button>',
                '        </div>',
                '      </div>',
                '    </div>',
                '  </div>',
                '  <!-- Alarm Config Subpage -->',
                '  <div class="settings-subpage" id="settings-alarm-config-subpage" style="display:none;">',
                '    <div class="settings-section">',
                '      <div class="settings-form">',
                '        <div class="form-group">',
                '          <label for="alarm-entry-delay" class="form-label">Entry Delay (seconds)</label>',
                '          <input type="number" id="alarm-entry-delay" class="form-input" min="0" max="600" placeholder="30" />',
                '          <div class="form-help">Seconds allowed to disarm after entry before alarm triggers.</div>',
                '        </div>',
                '        <div class="form-group">',
                '          <label for="alarm-exit-delay" class="form-label">Exit Delay (seconds)</label>',
                '          <input type="number" id="alarm-exit-delay" class="form-input" min="0" max="600" placeholder="30" />',
                '          <div class="form-help">Seconds allowed to leave after arming before alarm becomes active.</div>',
                '        </div>',
                '        <div class="form-status">',
                '          <span class="status-label">Current Settings:</span>',
                '          <span class="status-value" id="alarm-config-status">Loading...</span>',
                '        </div>',
                '        <div class="form-actions">',
                '          <button class="settings-save-btn" id="alarm-config-save-btn">Save</button>',
                '        </div>',
                '      </div>',
                '    </div>',
                '  </div>',
                '  <div class="settings-error" id="settings-error" style="display:none;"></div>',
                '</div>'
            ].join('\n');
            
            container.appendChild(viewElement);
            
            // Setup event listeners
            this._setupEventListeners();
            
            // Show main menu by default
            this.currentSubpage = 'main';
            this._showSubpage('main');
            
            // Load current HA state and update form
            if (window.SmartDisplay.store) {
                var state = window.SmartDisplay.store.getState();
                this.update(state);
            }

            // FAZ L4: Check advisor when entering settings
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
                    currentView: 'settings'
                });
            }

            // Schedule menu auto-hide timer (10s idle, 5m full close)
            if (window.SmartDisplay.viewManager) {
                var homeView = window.SmartDisplay.viewManager.views['home'];
                if (homeView && homeView._scheduleMenuAutoHide) {
                    console.log('[SettingsView] Scheduling menu auto-hide timer on mount');
                    homeView._scheduleMenuAutoHide();
                }
            }
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
            
            // Update HA status display
            if (data.haState) {
                var statusEl = document.getElementById('ha-status');
                if (statusEl) {
                    var statusText = '';
                    if (!data.haState.isConfigured) {
                        statusText = 'Not configured';
                    } else if (!data.haState.isConnected) {
                        statusText = 'Configured but not connected';
                    } else if (!data.haState.syncDone) {
                        statusText = 'Connected (sync pending)';
                    } else {
                        statusText = 'Connected and synced';
                        if (data.haState.entityCounts) {
                            var counts = data.haState.entityCounts;
                            var total = (counts.lights || 0) + (counts.sensors || 0) + 
                                       (counts.switches || 0) + (counts.others || 0);
                            statusText += ' (' + total + ' entities)';
                        }
                    }
                    statusEl.textContent = statusText;
                }
                
                // Auto-populate form fields with saved values
                if (data.haState.isConfigured) {
                    var serverInput = document.getElementById('ha-server-addr');
                    var tokenInput = document.getElementById('ha-token');
                    
                    // Set server URL from database if available
                    if (serverInput && data.haState.server_url) {
                        serverInput.value = data.haState.server_url;
                        serverInput.placeholder = 'Enter HA server address';
                    } else if (serverInput && !serverInput.value) {
                        serverInput.placeholder = 'Already configured (enter new address to update)';
                    }
                    
                    // Token is secure - show placeholder indicating it's configured
                    if (tokenInput && !tokenInput.value) {
                        tokenInput.placeholder = 'Already configured (enter new token to update)';
                    }
                } else {
                    // Reset placeholders for unconfigured state
                    var serverInput = document.getElementById('ha-server-addr');
                    var tokenInput = document.getElementById('ha-token');
                    if (serverInput) {
                        serverInput.placeholder = 'http://homeassistant.local:8123';
                    }
                    if (tokenInput) {
                        tokenInput.placeholder = '••••••••••••••••••••';
                    }
                }
            }
        },

        // ====================================================================
        // Private Methods
        // ====================================================================

        _setupEventListeners: function() {
            var self = this;

            // Back button - return to Home
            var backBtn = document.getElementById('settings-back-btn');
            if (backBtn) {
                backBtn.addEventListener('click', function() {
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    // Navigate to home view
                    if (window.SmartDisplay.store) {
                        window.SmartDisplay.store.setState({
                            menu: {
                                currentView: 'home',
                                isOpen: false
                            }
                        });
                    }
                });
            }

            // Home Assistant settings button
            var haBtn = document.getElementById('settings-ha-btn');
            if (haBtn) {
                haBtn.addEventListener('click', function() {
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    self._showSubpage('ha-settings');
                });
            }

            // HA Save button
            var saveBtn = document.getElementById('ha-save-btn');
            if (saveBtn) {
                saveBtn.addEventListener('click', function() {
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    self._handleHASave();
                });
            }

            // HA Test button
            var testBtn = document.getElementById('ha-test-btn');
            if (testBtn) {
                testBtn.addEventListener('click', function() {
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    self._handleHATest();
                });
            }

            // HA Sync button
            var syncBtn = document.getElementById('ha-sync-btn');
            if (syncBtn) {
                syncBtn.addEventListener('click', function() {
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    self._handleHASync();
                });
            }

            // Alarmo Monitoring button
            var alarmoBtn = document.getElementById('settings-alarmo-btn');
            if (alarmoBtn) {
                alarmoBtn.addEventListener('click', function() {
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    console.log('[SettingsView] Alarmo Monitoring button clicked');
                    ViewManager.routeToView('alarmo-settings');
                });
            }

            // Alarm Config button
            var alarmConfigBtn = document.getElementById('settings-alarm-config-btn');
            if (alarmConfigBtn) {
                alarmConfigBtn.addEventListener('click', function() {
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    self._showSubpage('alarm-config');
                    self._loadAlarmConfig();
                });
            }

            // Alarm Config Save button
            var alarmSaveBtn = document.getElementById('alarm-config-save-btn');
            if (alarmSaveBtn) {
                alarmSaveBtn.addEventListener('click', function() {
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    self._handleAlarmConfigSave();
                });
            }
        },

        _showSubpage: function(subpageName) {
            var mainEl = document.getElementById('settings-main');
            var haEl = document.getElementById('settings-ha-subpage');
            var alarmConfigEl = document.getElementById('settings-alarm-config-subpage');
            var backBtn = document.getElementById('settings-back-btn');

            if (subpageName === 'main') {
                if (mainEl) mainEl.style.display = 'block';
                if (haEl) haEl.style.display = 'none';
                if (alarmConfigEl) alarmConfigEl.style.display = 'none';
                if (backBtn) backBtn.style.display = 'none';
                this.currentSubpage = 'main';
            } else if (subpageName === 'ha-settings') {
                if (mainEl) mainEl.style.display = 'none';
                if (haEl) haEl.style.display = 'block';
                if (alarmConfigEl) alarmConfigEl.style.display = 'none';
                if (backBtn) backBtn.style.display = 'block';
                this.currentSubpage = 'ha-settings';
                console.log('[SettingsView] Showing HA Settings subpage');
            } else if (subpageName === 'alarm-config') {
                if (mainEl) mainEl.style.display = 'none';
                if (haEl) haEl.style.display = 'none';
                if (alarmConfigEl) alarmConfigEl.style.display = 'block';
                if (backBtn) backBtn.style.display = 'block';
                this.currentSubpage = 'alarm-config';
                console.log('[SettingsView] Showing Alarm Config subpage');
            }
        },

        _goBackToMain: function() {
            console.log('[SettingsView] Going back to main menu');
            this._showSubpage('main');
        },

        _loadAlarmConfig: function() {
            var self = this;
            console.log('[SettingsView] Loading alarm config...');
            
            window.SmartDisplay.api.client.get('/ui/settings', {
                headers: {
                    'X-User-Role': 'admin'
                }
            })
                .then(function(response) {
                    if (response && response.response && response.response.data && 
                        response.response.data.sections && response.response.data.sections.security) {
                        
                        var securityFields = response.response.data.sections.security.fields || [];
                        var entryDelay = 30, exitDelay = 30;
                        
                        securityFields.forEach(function(field) {
                            if (field.id === 'alarm_entry_delay_s') {
                                entryDelay = field.value || field.defaultValue || 30;
                            } else if (field.id === 'alarm_exit_delay_s') {
                                exitDelay = field.value || field.defaultValue || 30;
                            }
                        });
                        
                        // Update form inputs
                        var entryInput = document.getElementById('alarm-entry-delay');
                        var exitInput = document.getElementById('alarm-exit-delay');
                        var statusEl = document.getElementById('alarm-config-status');
                        
                        if (entryInput) entryInput.value = entryDelay;
                        if (exitInput) exitInput.value = exitDelay;
                        if (statusEl) {
                            statusEl.textContent = 'Entry: ' + entryDelay + 's, Exit: ' + exitDelay + 's';
                        }
                        
                        console.log('[SettingsView] Loaded alarm config - entry:', entryDelay, 'exit:', exitDelay);
                    }
                })
                .catch(function(error) {
                    console.error('[SettingsView] Failed to load alarm config:', error);
                    self._showError('Failed to load alarm settings');
                });
        },

        _handleAlarmConfigSave: function() {
            var self = this;
            var entryInput = document.getElementById('alarm-entry-delay');
            var exitInput = document.getElementById('alarm-exit-delay');
            
            if (!entryInput || !exitInput) {
                this._showError('Input fields not found');
                return;
            }
            
            var entryDelay = parseInt(entryInput.value) || 30;
            var exitDelay = parseInt(exitInput.value) || 30;
            
            console.log('[SettingsView] Saving alarm config - entry:', entryDelay, 'exit:', exitDelay);
            this._showStatus('Saving...');
            
            // Save entry delay
            window.SmartDisplay.api.client.post('/ui/settings/action', {
                action: 'field_change',
                field_id: 'alarm_entry_delay_s',
                new_value: entryDelay,
                confirm: true
            }, {
                headers: {
                    'X-User-Role': 'admin'
                }
            })
            .then(function() {
                // Save exit delay
                return window.SmartDisplay.api.client.post('/ui/settings/action', {
                    action: 'field_change',
                    field_id: 'alarm_exit_delay_s',
                    new_value: exitDelay,
                    confirm: true
                }, {
                    headers: {
                        'X-User-Role': 'admin'
                    }
                });
            })
            .then(function() {
                self._showStatus('Saved successfully');
                var statusEl = document.getElementById('alarm-config-status');
                if (statusEl) {
                    statusEl.textContent = 'Entry: ' + entryDelay + 's, Exit: ' + exitDelay + 's';
                }
                console.log('[SettingsView] Alarm config saved successfully');
            })
            .catch(function(error) {
                console.error('[SettingsView] Failed to save alarm config:', error);
                self._showError('Failed to save alarm settings');
            });
        },

        _handleHASave: function() {
            var self = this;
            var serverAddr = document.getElementById('ha-server-addr');
            var token = document.getElementById('ha-token');

            console.log('[SettingsView] HA Save button clicked');
            console.log('[SettingsView] Server input element:', serverAddr);
            console.log('[SettingsView] Token input element:', token);

            if (!serverAddr || !token) {
                this._showError('Input fields not found');
                return;
            }

            var serverUrl = serverAddr.value.trim();
            var tokenVal = token.value.trim();

            console.log('[SettingsView] Server URL value:', serverUrl);
            console.log('[SettingsView] Token value (length):', tokenVal.length);

            if (!serverUrl || !tokenVal) {
                this._showError('Server address and token are required');
                return;
            }

            // Show saving status
            this._showStatus('Saving...');

            // Call settings controller to save credentials
            window.SmartDisplay.settings.saveCredentials(serverUrl, tokenVal)
                .then(function(response) {
                    self._showStatus('Saved successfully - refreshing sensors...');
                    console.log('[SettingsView] HA credentials saved, refreshing alarmo sensors');
                    
                    // Update form fields with the successfully saved values
                    // Use server URL from response if available, otherwise use user input
                    var savedServerUrl = (response && response.server_url) ? response.server_url : serverUrl;
                    var serverInput = document.getElementById('ha-server-addr');
                    var tokenInput = document.getElementById('ha-token');
                    if (serverInput) {
                        serverInput.value = savedServerUrl; // Keep the server URL that was just saved
                        serverInput.placeholder = 'Server address saved successfully';
                    }
                    if (tokenInput) {
                        tokenInput.value = ''; // Clear token for security
                        tokenInput.placeholder = 'Token saved successfully (hidden for security)';
                    }
                    
                    // FAZ L6: Add trace entry
                    if (window.SmartDisplay.trace) {
                        window.SmartDisplay.trace.add('HA settings saved via UI');
                    }
                    
                    // Refresh HA status to show updated connection state
                    if (window.SmartDisplay.settings && window.SmartDisplay.settings.fetchHAStatus) {
                        window.SmartDisplay.settings.fetchHAStatus()
                            .then(function(haStatusUpdates) {
                                if (window.SmartDisplay.store && haStatusUpdates) {
                                    window.SmartDisplay.store.setState(haStatusUpdates);
                                    // Trigger update of this view with new HA state
                                    self.update(window.SmartDisplay.store.getState());
                                }
                            })
                            .catch(function(err) {
                                console.log('[SettingsView] HA status refresh after save failed:', err);
                            });
                    }
                    
                    // Automatically refresh alarmo sensors after successful HA config save
                    // Wait 1 second to allow backend alarmo adapter to reinitialize
                    setTimeout(function() {
                        console.log('[SettingsView] Triggering alarmo sensor refresh after HA config update');
                        // Trigger a global sensor refresh by posting to store
                        if (window.SmartDisplay.api) {
                            window.SmartDisplay.api.client.get('/ui/alarmo/sensors', {
                                headers: {
                                    'X-User-Role': (window.SmartDisplay.store.getState().authState || {}).role || 'admin'
                                }
                            })
                            .then(function(envelope) {
                                var response = envelope.response || {};
                                if (response.ok && response.data) {
                                    // Update global store with new sensors
                                    window.SmartDisplay.store.setState({
                                        alarmoState: {
                                            sensors: Array.isArray(response.data) ? response.data : []
                                        }
                                    });
                                    console.log('[SettingsView] Alarmo sensors updated in store after HA config save');
                                }
                            })
                            .catch(function(err) {
                                console.log('[SettingsView] Sensor refresh after HA save failed (non-critical):', err);
                            });
                        }
                    }, 1000);
                })
                .catch(function(err) {
                    self._showError(err.message || 'Failed to save credentials');
                    console.error('[SettingsView] HA save error:', err);
                });
        },

        _handleHATest: function() {
            var self = this;
            console.log('[SettingsView] HA Test button clicked');

            this._showStatus('Testing connection...');

            window.SmartDisplay.settings.testHAConnection()
                .then(function() {
                    self._showStatus('Connection successful!');
                    console.log('[SettingsView] HA test successful');
                })
                .catch(function(err) {
                    self._showError(err.message || 'Connection test failed');
                    console.error('[SettingsView] HA test error:', err);
                });
        },

        _handleHASync: function() {
            var self = this;
            console.log('[SettingsView] HA Sync button clicked');

            this._showStatus('Syncing data from HA...');

            window.SmartDisplay.settings.performSync()
                .then(function() {
                    self._showStatus('Sync completed successfully!');
                    console.log('[SettingsView] HA sync successful');
                })
                .catch(function(err) {
                    self._showError(err.message || 'Sync failed');
                    console.error('[SettingsView] HA sync error:', err);
                });
        },

        _showStatus: function(message) {
            var statusEl = document.getElementById('ha-status');
            if (statusEl) {
                statusEl.textContent = message;
            }
        },

        _showError: function(message) {
            var errorEl = document.getElementById('settings-error');
            if (errorEl) {
                errorEl.textContent = 'Error: ' + message;
                errorEl.style.display = 'block';
            }
        },

        _clearError: function() {
            var errorEl = document.getElementById('settings-error');
            if (errorEl) {
                errorEl.style.display = 'none';
                errorEl.textContent = '';
            }
        }
    };

    /**
     * Alarmo Settings View
     * Read-only monitoring of Home Assistant Alarmo status, sensors, and events
     * Mounted at #/settings/homeassistant/alarmo
     */
    var AlarmoSettingsView = {
        id: 'alarmo-settings',
        name: 'Alarmo Monitoring',
        pollingHandle: null,
        currentFilter: 'all',

        mount: function() {
            console.log('[View] Mounting AlarmoSettingsView');
            var container = document.getElementById('app');
            
            container.innerHTML = '';
            var viewElement = document.createElement('div');
            viewElement.id = this.id;
            viewElement.className = 'view view-alarmo-settings';
            
            viewElement.innerHTML = [
                '<div class="alarmo-header">',
                '  <h1 class="alarmo-title">Alarmo Monitoring</h1>',
                '  <button class="alarmo-back-btn" id="alarmo-back-btn">← Back</button>',
                '</div>',
                '<div class="alarmo-container">',
                '  <!-- Status Panel -->',
                '  <div class="alarmo-status-panel" id="alarmo-status-panel">',
                '    <div class="status-box">',
                '      <div class="status-indicator" id="status-indicator"></div>',
                '      <div class="status-info">',
                '        <p class="status-label">Connection Status</p>',
                '        <p class="status-value" id="status-value">Loading...</p>',
                '        <p class="status-timestamp" id="status-timestamp"></p>',
                '      </div>',
                '    </div>',
                '    <p class="status-note">Alarm logic is managed by Home Assistant (read-only)</p>',
                '  </div>',
                '  <!-- Sensor Cards Grid -->',
                '  <div class="alarmo-section">',
                '    <h2 class="section-title">Sensor Status</h2>',
                '    <div class="sensor-grid" id="sensor-grid">',
                '      <p class="placeholder">Loading sensors...</p>',
                '    </div>',
                '  </div>',
                '  <!-- Event Log -->',
                '  <div class="alarmo-section">',
                '    <h2 class="section-title">Recent Alarmo Activity</h2>',
                '    <div class="event-filters" id="event-filters" style="display:none;">',
                '      <button class="filter-btn filter-btn-active" data-filter="all">All</button>',
                '      <button class="filter-btn" data-filter="sensors">Sensors</button>',
                '      <button class="filter-btn" data-filter="alarm">Alarm</button>',
                '      <button class="filter-btn" data-filter="system">System</button>',
                '    </div>',
                '    <div class="event-list" id="event-list">',
                '      <p class="placeholder">Loading activity...</p>',
                '    </div>',
                '    <div class="event-actions">',
                '      <button class="refresh-btn" id="refresh-events-btn">Refresh activity</button>',
                '    </div>',
                '  </div>',
                '  <div class="alarmo-error" id="alarmo-error" style="display:none;"></div>',
                '</div>'
            ].join('\n');
            
            container.appendChild(viewElement);
            
            // Setup event listeners
            this._setupEventListeners();
            
            // Schedule menu auto-hide timer (10s idle, 5m full close)
            if (window.SmartDisplay.viewManager) {
                var homeView = window.SmartDisplay.viewManager.views['home'];
                if (homeView && homeView._scheduleMenuAutoHide) {
                    console.log('[AlarmoSettingsView] Scheduling menu auto-hide timer on mount');
                    homeView._scheduleMenuAutoHide();
                }
            }
            
            // Fetch data
            this.refresh();
            
            // Setup polling
            this.setupPolling();
        },

        unmount: function() {
            console.log('[View] Unmounting AlarmoSettingsView');
            if (this.pollingHandle) {
                clearInterval(this.pollingHandle);
                this.pollingHandle = null;
            }
            var element = document.getElementById(this.id);
            if (element) {
                element.remove();
            }
        },

        update: function(data) {
            console.log('[View] Updating AlarmoSettingsView', data);
            if (data && data.alarmoState) {
                this.renderStatus(data.alarmoState.status);
                this.renderSensors(data.alarmoState.sensors);
                this.renderEvents(data.alarmoState.events);
            }
        },

        _setupEventListeners: function() {
            var self = this;

            // Back button
            var backBtn = document.getElementById('alarmo-back-btn');
            if (backBtn) {
                backBtn.addEventListener('click', function() {
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    ViewManager.routeToView('settings');
                });
            }

            // Filter buttons
            var filterBtns = document.querySelectorAll('.filter-btn');
            filterBtns.forEach(function(btn) {
                btn.addEventListener('click', function() {
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    self.setFilter(this.getAttribute('data-filter'));
                });
            });

            // Refresh button
            var refreshBtn = document.getElementById('refresh-events-btn');
            if (refreshBtn) {
                refreshBtn.addEventListener('click', function() {
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    self.refreshEvents();
                });
            }
        },

        setupPolling: function() {
            var self = this;
            // Poll every 30 seconds
            this.pollingHandle = setInterval(function() {
                self.refresh();
            }, 30000);
        },

        refresh: function() {
            var self = this;
            var role = (window.SmartDisplay.store.getState().authState || {}).role || 'guest';

            console.log('[AlarmoSettingsView] Refreshing data...');

            // Always fetch status
            this.fetchStatus();

            // Always fetch sensors
            this.fetchSensors();

            // Only fetch events for admin/user
            if (role !== 'guest') {
                this.fetchEvents();
            }
        },

        refreshEvents: function() {
            var self = this;
            var role = (window.SmartDisplay.store.getState().authState || {}).role || 'guest';

            if (role === 'guest') {
                this._showError('Guests cannot view event log');
                return;
            }

            // Show loading state
            var refreshBtn = document.getElementById('refresh-events-btn');
            if (refreshBtn) {
                refreshBtn.disabled = true;
                refreshBtn.textContent = 'Loading...';
            }

            this.fetchEvents()
                .then(function() {
                    if (refreshBtn) {
                        refreshBtn.disabled = false;
                        refreshBtn.textContent = 'Refresh activity';
                    }
                })
                .catch(function(err) {
                    console.error('[AlarmoSettingsView] Refresh failed:', err);
                    self._showError('Failed to refresh activity');
                    if (refreshBtn) {
                        refreshBtn.disabled = false;
                        refreshBtn.textContent = 'Refresh activity';
                    }
                });
        },

        fetchStatus: function() {
            var self = this;
            console.log('[AlarmoSettingsView] Fetching status...');
            return window.SmartDisplay.api.client.get('/ui/alarmo/status', {
                headers: {
                    'X-User-Role': (window.SmartDisplay.store.getState().authState || {}).role || 'guest'
                }
            })
            .then(function(envelope) {
                console.log('[AlarmoSettingsView] Status response:', envelope);
                var response = envelope.response || {};
                if (!response.ok) {
                    throw new Error(response.error || 'API error');
                }

                var data = response.data || {};
                console.log('[AlarmoSettingsView] Status data:', data);

                // Derive best-effort mode from HA fields for accurate UI after refresh
                var resolvedState = data.alarmo_state || data.alarmo_raw_state || '';
                if (!resolvedState && data.alarmo_mode === 'armed' && data.alarmo_armed_mode) {
                    resolvedState = 'armed_' + data.alarmo_armed_mode;
                } else if (!resolvedState && data.alarmo_mode) {
                    resolvedState = data.alarmo_mode;
                }

                var delayRemaining = typeof data.delay_remaining === 'number' ? data.delay_remaining : 0;

                window.SmartDisplay.store.setState({
                    alarmoState: {
                        status: {
                            alarmo_connected: data.alarmo_connected || false,
                            ha_runtime_unreachable: data.ha_runtime_unreachable || false,
                            last_seen_at: data.last_seen_at || null,
                            alarmo_state: resolvedState || null,
                            alarmo_mode: data.alarmo_mode || null,
                            alarmo_armed_mode: data.alarmo_armed_mode || null,
                            alarmo_raw_state: data.alarmo_raw_state || null,
                            alarmo_triggered: data.alarmo_triggered === true,
                            delay_remaining: delayRemaining,
                            delay_type: data.delay_type || '',
                            alarmo_last_changed: data.alarmo_last_changed || null
                        },
                        alarmo_state: resolvedState || null,
                        alarmo_triggered: data.alarmo_triggered === true,
                        delay_remaining: delayRemaining
                    },
                    alarmoControl: resolvedState ? { currentMode: resolvedState } : {}
                });

                self.renderStatus(window.SmartDisplay.store.getState().alarmoState.status);
                self._clearError();
            })
            .catch(function(err) {
                console.error('[AlarmoSettingsView] Status fetch error:', err);
                self._showError('Failed to fetch Alarmo status: ' + (err.message || err));
            });
        },

        fetchSensors: function() {
            var self = this;
            console.log('[AlarmoSettingsView] Fetching sensors...');
            return window.SmartDisplay.api.client.get('/ui/alarmo/sensors', {
                headers: {
                    'X-User-Role': (window.SmartDisplay.store.getState().authState || {}).role || 'guest'
                }
            })
            .then(function(envelope) {
                console.log('[AlarmoSettingsView] Sensors response:', envelope);
                var response = envelope.response || {};
                if (!response.ok) {
                    throw new Error(response.error || 'API error');
                }

                var data = response.data || [];
                console.log('[AlarmoSettingsView] Sensors data:', data);
                window.SmartDisplay.store.setState({
                    alarmoState: {
                        sensors: Array.isArray(data) ? data : []
                    }
                });

                self.renderSensors(data);
            })
            .catch(function(err) {
                console.error('[AlarmoSettingsView] Sensors fetch error:', err);
                // Soft fail - show empty grid
                self.renderSensors([]);
            });
        },

        fetchEvents: function() {
            var self = this;
            var role = (window.SmartDisplay.store.getState().authState || {}).role || 'guest';

            if (role === 'guest') {
                console.log('[AlarmoSettingsView] Guest user - skipping event fetch');
                window.SmartDisplay.store.setState({
                    alarmoState: {
                        events: []
                    }
                });
                return Promise.resolve();
            }

            console.log('[AlarmoSettingsView] Fetching events...');
            return window.SmartDisplay.api.client.get('/ui/alarmo/events?limit=20', {
                headers: {
                    'X-User-Role': role
                }
            })
            .then(function(envelope) {
                console.log('[AlarmoSettingsView] Events response:', envelope);
                var response = envelope.response || {};
                if (!response.ok) {
                    throw new Error(response.error || 'API error');
                }

                var data = response.data || [];
                console.log('[AlarmoSettingsView] Events data:', data);
                window.SmartDisplay.store.setState({
                    alarmoState: {
                        events: Array.isArray(data) ? data : []
                    }
                });

                self.renderEvents(data);
                self._clearError();
            })
            .catch(function(err) {
                console.error('[AlarmoSettingsView] Events fetch error:', err);
                // Soft fail - show empty list
                self.renderEvents([]);
            });
        },

        renderStatus: function(status) {
            if (!status) return;

            var indicator = document.getElementById('status-indicator');
            var valueEl = document.getElementById('status-value');
            var timestampEl = document.getElementById('status-timestamp');

            if (!indicator || !valueEl) return;

            var isConnected = status.alarmo_connected && !status.ha_runtime_unreachable;
            var statusText = '';
            var statusClass = '';

            if (!status.alarmo_connected) {
                statusText = 'Disconnected';
                statusClass = 'status-error';
            } else if (status.ha_runtime_unreachable) {
                statusText = 'Connected (Runtime Unreachable)';
                statusClass = 'status-warning';
            } else {
                statusText = 'Connected & Healthy';
                statusClass = 'status-ok';
            }

            indicator.className = 'status-indicator ' + statusClass;
            valueEl.textContent = statusText;

            if (status.last_seen_at && timestampEl) {
                var lastSeen = new Date(status.last_seen_at);
                var now = new Date();
                var diffSec = Math.floor((now - lastSeen) / 1000);

                var timeStr = '';
                if (diffSec < 60) {
                    timeStr = 'Just now';
                } else if (diffSec < 3600) {
                    timeStr = Math.floor(diffSec / 60) + ' min ago';
                } else if (diffSec < 86400) {
                    timeStr = Math.floor(diffSec / 3600) + ' h ago';
                } else {
                    timeStr = Math.floor(diffSec / 86400) + ' d ago';
                }

                timestampEl.textContent = 'Last seen: ' + timeStr;
            }
        },

        renderSensors: function(sensors) {
            var grid = document.getElementById('sensor-grid');
            if (!grid) return;

            if (!sensors || sensors.length === 0) {
                grid.innerHTML = '<p class="placeholder">No sensors available</p>';
                return;
            }

            var html = sensors.map(function(sensor) {
                var deviceClass = sensor.device_class || 'sensor';
                var stateLabel = sensor.state === 'on' || sensor.state === 'open' 
                    ? 'Active' 
                    : 'Inactive';
                
                var healthClass = 'health-ok';
                // TODO: Add battery check and 24h check when available

                return [
                    '<div class="sensor-card">',
                    '  <div class="sensor-header">',
                    '    <h3 class="sensor-name">' + (sensor.name || 'Unknown') + '</h3>',
                    '    <span class="health-indicator health-' + healthClass + '"></span>',
                    '  </div>',
                    '  <div class="sensor-body">',
                    '    <p class="sensor-type">' + deviceClass + '</p>',
                    '    <p class="sensor-state">' + stateLabel + '</p>',
                    '  </div>',
                    '  <div class="sensor-footer">',
                    '    <p class="sensor-timestamp">' + (sensor.last_changed ? new Date(sensor.last_changed).toLocaleString() : 'N/A') + '</p>',
                    '  </div>',
                    '</div>'
                ].join('\n');
            }).join('\n');

            grid.innerHTML = html;
        },

        renderEvents: function(events) {
            var role = (window.SmartDisplay.store.getState().authState || {}).role || 'guest';
            var listEl = document.getElementById('event-list');
            var filtersEl = document.getElementById('event-filters');

            if (!listEl) return;

            // Hide event log for guests
            if (role === 'guest') {
                if (filtersEl) filtersEl.style.display = 'none';
                listEl.innerHTML = '<p class="placeholder">Event log not available for guests</p>';
                return;
            }

            // Show filters for admin/user
            if (filtersEl) filtersEl.style.display = 'flex';

            if (!events || events.length === 0) {
                listEl.innerHTML = '<p class="placeholder">No recent activity</p>';
                return;
            }

            // Filter events
            var filtered = events;
            if (this.currentFilter !== 'all') {
                filtered = events.filter(function(evt) {
                    return evt.event_type === this.currentFilter;
                }, this);
            }

            if (filtered.length === 0) {
                listEl.innerHTML = '<p class="placeholder">No events match filter</p>';
                return;
            }

            var html = filtered.map(function(event) {
                var timestamp = new Date(event.last_changed).toLocaleString();
                var chipClass = 'event-chip-' + (event.event_type || 'info');

                return [
                    '<div class="event-item">',
                    '  <div class="event-time">' + timestamp + '</div>',
                    '  <div class="event-entity">' + (event.name || event.entity_id || 'Unknown') + '</div>',
                    '  <div class="event-chip ' + chipClass + '">' + event.state + '</div>',
                    '</div>'
                ].join('\n');
            }).join('\n');

            listEl.innerHTML = html;
        },

        setFilter: function(filter) {
            this.currentFilter = filter;

            // Update button states
            var filterBtns = document.querySelectorAll('.filter-btn');
            filterBtns.forEach(function(btn) {
                if (btn.getAttribute('data-filter') === filter) {
                    btn.classList.add('filter-btn-active');
                } else {
                    btn.classList.remove('filter-btn-active');
                }
            });

            // Re-render with filter
            var state = window.SmartDisplay.store.getState();
            this.renderEvents(state.alarmoState.events);
        },

        _showError: function(message) {
            var errorEl = document.getElementById('alarmo-error');
            if (errorEl) {
                errorEl.textContent = message;
                errorEl.style.display = 'block';
            }
        },

        _clearError: function() {
            var errorEl = document.getElementById('alarmo-error');
            if (errorEl) {
                errorEl.style.display = 'none';
                errorEl.textContent = '';
            }
        }
    };

    /**
     * Alarm Control View
     * Premium glass glow design with modern PIN entry and mode selection
     * Completely rewritten for modern UX
     * 
     * Extended with Alarmo state-driven overlay system:
     * - Fullscreen blur + pulse for pending/arming/triggered
     * - Countdown driven ONLY by alarmo.delay.remaining
     * - Triggered state PIN overlay
     * - Role-based PIN visibility
     */
    var AlarmControlView = {
        id: 'alarm-control',
        name: 'Alarm Control',
        currentMode: null,
        pollingInterval: null,
        lastKnownState: null,
        countdownTimer: null, // Real-time countdown timer
        countdownStartTime: null, // When countdown started
        countdownDuration: 0, // Total countdown duration
        
        mount: function() {
            console.log('[AlarmControlView] Mounting premium glass glow design with Alarmo overlay system...');
            var mainEl = document.getElementById('main-content');
            if (!mainEl) return;

            mainEl.innerHTML = `
                <!-- Alarm Overlay System -->
                <div class="alarm-overlay" id="alarm-overlay"></div>
                
                <!-- Countdown Overlay (pending/arming) -->
                <div class="alarm-countdown-overlay" id="alarm-countdown-overlay">
                    <div class="alarm-countdown-label">Alarm will activate in</div>
                    <div class="alarm-countdown-value" id="alarm-countdown-value">--</div>
                </div>
                
                <!-- Triggered State PIN Overlay -->
                <div class="alarm-triggered-overlay" id="alarm-triggered-overlay">
                    <div class="triggered-header">
                        <div class="triggered-title">⚠️ ALARM TRIGGERED</div>
                        <div class="triggered-subtitle">Immediate action required</div>
                    </div>
                    
                    <div class="triggered-pin-section" id="triggered-pin-section">
                        <div class="triggered-pin-label">Enter PIN to disarm</div>
                        <input type="password" id="triggered-pin-input" class="triggered-pin-input" maxlength="6" placeholder="••••••" autocomplete="off">
                        
                        <div class="triggered-numpad">
                            <button class="triggered-numpad-btn" data-num="1">1</button>
                            <button class="triggered-numpad-btn" data-num="2">2</button>
                            <button class="triggered-numpad-btn" data-num="3">3</button>
                            <button class="triggered-numpad-btn" data-num="4">4</button>
                            <button class="triggered-numpad-btn" data-num="5">5</button>
                            <button class="triggered-numpad-btn" data-num="6">6</button>
                            <button class="triggered-numpad-btn" data-num="7">7</button>
                            <button class="triggered-numpad-btn" data-num="8">8</button>
                            <button class="triggered-numpad-btn" data-num="9">9</button>
                            <button class="triggered-numpad-btn" data-num="0">0</button>
                            <button class="triggered-numpad-btn clear" data-num="clear">⌫</button>
                        </div>
                        
                        <button class="triggered-disarm-btn" id="triggered-disarm-btn">Disarm Now</button>
                    </div>
                    
                    <div class="triggered-guest-warning" id="triggered-guest-warning" style="display: none;">
                        <p>🔒 Guest mode active<br>Only administrators can disarm the alarm</p>
                    </div>
                    
                    <div class="triggered-countdown" id="triggered-countdown" style="display: none;">
                        <div class="triggered-countdown-label">Alarm escalation in</div>
                        <div class="triggered-countdown-value" id="triggered-countdown-value">--</div>
                    </div>
                </div>
                
                <div class="alarm-container compact">
                    <!-- Durum Göstergesi -->
                    <div class="alarm-state-indicator">
                        <h3>Mevcut Durum</h3>
                        <div class="state-value" id="alarm-state-display">Yükleniyor...</div>
                    </div>

                    <!-- Mesaj Gösterimi -->
                    <div id="alarm-message" class="alarm-message"></div>

                    <!-- Mod Seçim Paneli -->
                    <div class="alarm-modes-section">
                        <label class="alarm-modes-label">Güvenlik Modu Seçiniz</label>
                        <div class="alarm-modes">
                            <button class="alarm-mode-btn" data-mode="disarmed">
                                <div class="mode-icon">🛡️</div>
                                <div class="mode-label">Devre Dışı</div>
                            </button>
                            <button class="alarm-mode-btn" data-mode="armed_away">
                                <div class="mode-icon">🔒</div>
                                <div class="mode-label">Dışarıda</div>
                            </button>
                            <button class="alarm-mode-btn" data-mode="armed_home">
                                <div class="mode-icon">🏠</div>
                                <div class="mode-label">Evde</div>
                            </button>
                            <button class="alarm-mode-btn" data-mode="armed_night">
                                <div class="mode-icon">🌙</div>
                                <div class="mode-label">Gece</div>
                            </button>
                        </div>
                    </div>

                    <!-- Kimlik Doğrulama Paneli -->
                    <div class="alarm-code-section">
                        <label>Güvenlik Kodu</label>
                        <input type="password" id="alarm-code-input" class="alarm-code-input" maxlength="6" placeholder="••••••" autocomplete="off">
                        
                        <!-- Sayısal Tuş Takımı -->
                        <div class="alarm-numpad">
                            <button class="numpad-btn" data-num="1">1</button>
                            <button class="numpad-btn" data-num="2">2</button>
                            <button class="numpad-btn" data-num="3">3</button>
                            
                            <button class="numpad-btn" data-num="4">4</button>
                            <button class="numpad-btn" data-num="5">5</button>
                            <button class="numpad-btn" data-num="6">6</button>
                            
                            <button class="numpad-btn" data-num="7">7</button>
                            <button class="numpad-btn" data-num="8">8</button>
                            <button class="numpad-btn" data-num="9">9</button>
                            
                            <button class="numpad-btn" data-num="0">0</button>
                            <button class="numpad-btn clear" data-num="clear">⌫</button>
                        </div>
                    </div>

                    <!-- Sensor Kartları (PIN pad altında) -->
                    <div class="sensor-cards-section">
                        <div class="sensor-card-grid" id="sensor-card-grid">
                            <p class="placeholder">Sensörler yükleniyor...</p>
                        </div>
                    </div>
                </div>
            `;

            this._setupEventListeners();
            this._setupTriggeredOverlayListeners();
            this.refresh();
            this.fetchStatus();
            this.fetchSensors();
            this.setupPolling();
            
            // Subscribe to store updates for real-time status
            var self = this;
            this.storeUnsubscribe = window.SmartDisplay.store.subscribe(function() {
                console.log('[AlarmControlView] Store update received');
                self.refresh();
                self.renderSensors();
            });
        },

        unmount: function() {
            console.log('[AlarmControlView] Unmounting...');
            this.stopPolling();
            
            // Clear countdown timer
            if (this.countdownTimer) {
                clearInterval(this.countdownTimer);
                this.countdownTimer = null;
            }
            
            if (this.storeUnsubscribe) {
                this.storeUnsubscribe();
                this.storeUnsubscribe = null;
            }
            this.lastKnownState = null;
            var mainEl = document.getElementById('main-content');
            if (mainEl) {
                mainEl.innerHTML = '';
            }
        },

        update: function(state) {
            console.log('[AlarmControlView] State update received');
            // Component manages its own state through polling
        },

        _setupEventListeners: function() {
            var self = this;

            // PIN input - focus on load for immediate entry
            var codeInput = document.getElementById('alarm-code-input');
            if (codeInput) {
                // Auto-focus for PIN entry
                setTimeout(function() {
                    codeInput.focus();
                }, 100);
                
                // Keyboard support: numbers, backspace, enter
                codeInput.addEventListener('keydown', function(e) {
                    if (e.key >= '0' && e.key <= '9') {
                        // Allow numbers
                        return true;
                    } else if (e.key === 'Backspace') {
                        // Allow native backspace
                        return true;
                    } else if (e.key === 'Enter') {
                        // Submit: disarm if armed, submit mode if mode selected
                        e.preventDefault();
                        var state = window.SmartDisplay.store.getState();
                        var currentMode = (state.alarmoControl || {}).currentMode || 'disarmed';
                        
                        if (currentMode === 'disarmed') {
                            // Disarmed: require mode selection
                            if (self.currentMode) {
                                self._submitMode();
                            }
                        } else {
                            // Armed: disarm immediately
                            self._submitDisarmFromArmed();
                        }
                        return false;
                    } else {
                        e.preventDefault();
                    }
                });
                
                // Auto-submit on reaching 4 digits
                codeInput.addEventListener('input', function() {
                    if (this.value.length >= 4) {
                        var state = window.SmartDisplay.store.getState();
                        var currentMode = (state.alarmoControl || {}).currentMode || 'disarmed';
                        
                        console.log('[AlarmControlView] PIN 4+ digits entered. Current mode:', currentMode);
                        
                        if (currentMode === 'disarmed') {
                            // Disarmed mode: auto-submit only if mode button selected
                            if (self.currentMode) {
                                console.log('[AlarmControlView] Auto-submitting mode:', self.currentMode);
                                setTimeout(function() {
                                    self._submitMode();
                                }, 100);
                            }
                        } else {
                            // Armed state: auto-disarm with PIN
                            console.log('[AlarmControlView] Auto-disarming from armed state');
                            setTimeout(function() {
                                self._submitDisarmFromArmed();
                            }, 100);
                        }
                    }
                });
                
                // Prevent paste
                codeInput.addEventListener('paste', function(e) {
                    e.preventDefault();
                });
            }

            // Numpad input handling
            var numpadBtns = document.querySelectorAll('.numpad-btn');
            numpadBtns.forEach(function(btn) {
                btn.addEventListener('click', function() {
                    var num = this.getAttribute('data-num');
                    self._handleNumpadInput(num);
                });
            });

            // Mode selection buttons - only relevant when disarmed
            var modeBtns = document.querySelectorAll('.alarm-mode-btn');
            modeBtns.forEach(function(btn) {
                btn.addEventListener('click', function() {
                    var state = window.SmartDisplay.store.getState();
                    var currentMode = (state.alarmoControl || {}).currentMode || 'disarmed';
                    
                    // Only allow mode selection when disarmed
                    if (currentMode !== 'disarmed') {
                        console.log('[AlarmControlView] Cannot change mode while armed. Current mode:', currentMode);
                        self._showMessage('⚠️ Kurulu moddan çıkmak için PIN girin', 'error');
                        return;
                    }
                    
                    var mode = this.getAttribute('data-mode');
                    var codeInput = document.getElementById('alarm-code-input');
                    var code = codeInput ? codeInput.value : '';
                    
                    console.log('[AlarmControlView] Mode button clicked:', mode, 'PIN length:', code.length);
                    
                    // Validate PIN before proceeding
                    if (!code || code.length < 4) {
                        self._showMessage('PIN gerekli (4 hane)', 'error');
                        if (codeInput) codeInput.focus();
                        return;
                    }
                    
                    // Visual feedback
                    modeBtns.forEach(function(b) { b.classList.remove('selected'); });
                    this.classList.add('selected');
                    
                    self.currentMode = mode;
                    self._clearMessage();
                    
                    // Submit mode with the PIN that was entered
                    console.log('[AlarmControlView] Submitting mode:', mode, 'with code');
                    self._submitMode();
                });
            });
        },

        _setupTriggeredOverlayListeners: function() {
            var self = this;

            // Triggered numpad buttons
            var triggeredNumpadBtns = document.querySelectorAll('.triggered-numpad-btn');
            triggeredNumpadBtns.forEach(function(btn) {
                btn.addEventListener('click', function() {
                    var num = this.getAttribute('data-num');
                    self._handleTriggeredNumpadInput(num);
                });
            });

            // Triggered PIN input - auto-disarm when 4+ digits entered
            var triggeredPinInput = document.getElementById('triggered-pin-input');
            if (triggeredPinInput) {
                // Keyboard support
                triggeredPinInput.addEventListener('keydown', function(e) {
                    if (e.key === 'Enter') {
                        e.preventDefault();
                        self._submitTriggeredDisarm();
                    } else if (e.key === 'Backspace') {
                        return true;
                    } else if (e.key >= '0' && e.key <= '9') {
                        return true;
                    } else {
                        e.preventDefault();
                    }
                });

                // Auto-disarm on reaching 4 digits
                triggeredPinInput.addEventListener('input', function() {
                    if (this.value.length >= 4) {
                        console.log('[AlarmControlView] Auto-disarming with PIN length:', this.value.length);
                        setTimeout(function() {
                            self._submitTriggeredDisarm();
                        }, 100);
                    }
                });

                triggeredPinInput.addEventListener('paste', function(e) {
                    e.preventDefault();
                });
            }
        },

        _handleTriggeredNumpadInput: function(num) {
            var input = document.getElementById('triggered-pin-input');
            if (!input) return;

            if (num === 'clear') {
                input.value = '';
                console.log('[AlarmControlView] Triggered PIN cleared');
            } else if (input.value.length < 6) {
                input.value += num;
                console.log('[AlarmControlView] Triggered PIN digit added, length:', input.value.length);
            }
        },

        _submitTriggeredDisarm: function() {
            var code = document.getElementById('triggered-pin-input').value.trim();

            if (!code || code === '') {
                console.log('[AlarmControlView] Triggered disarm: no PIN provided');
                return;
            }

            if (code.length < 4) {
                console.log('[AlarmControlView] Triggered disarm: PIN too short');
                return;
            }

            console.log('[AlarmControlView] 🚨 Triggered disarm request, code length:', code.length);
            this._sendDisarm(code);
        },

        _handleNumpadInput: function(num) {
            var input = document.getElementById('alarm-code-input');
            if (!input) return;

            if (num === 'clear') {
                input.value = '';
                console.log('[AlarmControlView] PIN cleared via numpad');
            } else if (input.value.length < 6) {
                input.value += num;
                console.log('[AlarmControlView] PIN digit added, length:', input.value.length);
            }

            // Decide action based on current armed/disarmed state from store
            var state = window.SmartDisplay.store.getState();
            var currentMode = (state.alarmoControl || {}).currentMode || 'disarmed';

            if (input.value.length >= 4) {
                if (currentMode === 'disarmed') {
                    // Require a selected mode to arm
                    if (this.currentMode) {
                        console.log('[AlarmControlView] Auto-submitting mode via numpad:', this.currentMode);
                        setTimeout(function() {
                            this._submitMode();
                        }.bind(this), 150);
                    }
                } else {
                    // Armed: auto-disarm on PIN entry
                    console.log('[AlarmControlView] Auto-disarming via numpad, store mode:', currentMode);
                    setTimeout(function() {
                        this._submitDisarmFromArmed();
                    }.bind(this), 150);
                }
            }
        },

        _submitMode: function() {
            console.log('[AlarmControlView] Submit mode requested');
            
            // Validate mode selection
            if (!this.currentMode) {
                this._showMessage('⚠️ Lütfen bir mod seçiniz', 'error');
                return;
            }

            // Get PIN code
            var code = document.getElementById('alarm-code-input').value.trim();

            // Validate PIN
            if (!code || code === '') {
                this._showMessage('🔑 Güvenlik kodunu giriniz', 'error');
                return;
            }

            if (code.length < 4) {
                this._showMessage('⚠️ Kod en az 4 haneli olmalıdır', 'error');
                return;
            }

            console.log('[AlarmControlView] Validation passed, mode:', this.currentMode, 'code length:', code.length);

            // Execute mode change
            if (this.currentMode === 'disarmed') {
                this._sendDisarm(code);
            } else {
                this._sendArm(this.currentMode, code);
            }
        },

        // Disarm directly from armed state by PIN only (no button)
        _submitDisarmFromArmed: function() {
            var code = document.getElementById('alarm-code-input').value.trim();

            if (!code || code.length < 4) {
                this._showMessage('PIN gerekli (4 hane)', 'error');
                return;
            }

            console.log('[AlarmControlView] Disarm from armed state with PIN length:', code.length);
            this._clearMessage();
            this._sendDisarm(code);
        },

        _clearForm: function() {
            console.log('[AlarmControlView] Clearing form...');
            
            // Clear PIN input
            var codeInput = document.getElementById('alarm-code-input');
            if (codeInput) {
                codeInput.value = '';
            }
            
            // Clear mode selection
            this.currentMode = null;
            var modeBtns = document.querySelectorAll('.alarm-mode-btn');
            modeBtns.forEach(function(btn) { 
                btn.classList.remove('selected'); 
            });
            
            // Clear messages
            this._clearMessage();
        },

        _sendArm: function(mode, code) {
            var self = this;
            console.log('[AlarmControlView] 🔐 Arming alarm - Mode:', mode, 'Code provided:', !!code);

            var payload = { mode: mode };
            if (code) {
                payload.code = code;
            }

            this._showMessage('⏳ Alarm etkinleştiriliyor...', 'info');

            window.SmartDisplay.api.client.post('/ui/alarmo/arm', payload, {
                headers: {
                    'X-User-Role': (window.SmartDisplay.store.getState().authState || {}).role || 'guest'
                }
            })
            .then(function(envelope) {
                console.log('[AlarmControlView] ✅ Arm response:', envelope);
                var response = envelope.response || {};
                
                if (!response.ok) {
                    throw new Error(response.error || 'ARM işlemi başarısız');
                }
                
                // Update store with new mode
                window.SmartDisplay.store.setState({
                    alarmoControl: { currentMode: mode }
                });
                
                self._showMessage('✓ Alarm başarıyla etkinleştirildi', 'success');
                
                // COUNTDOWN BAŞLAT - Alarm kurulduktan sonra 30 saniye delay
                console.log('[AlarmControlView] 🚀 Alarm kuruldu! Countdown başlatılıyor (30 saniye)...');
                self._startNewCountdown(30);
                
                // Form temizlemeyi geciktir (refresh yapma, countdown'u bozar)
                setTimeout(function() {
                    self._clearForm();
                }, 2000);
            })
            .catch(function(err) {
                console.error('[AlarmControlView] ❌ Arm error:', err);
                self._showMessage('✗ Hata: ' + (err.message || err), 'error');
            });
        },

        _sendDisarm: function(code) {
            var self = this;
            console.log('[AlarmControlView] 🔓 Disarming alarm - Code provided:', !!code);

            var payload = {};
            if (code) {
                payload.code = code;
            }

            this._showMessage('⏳ Alarm devre dışı bırakılıyor...', 'info');

            window.SmartDisplay.api.client.post('/ui/alarmo/disarm', payload, {
                headers: {
                    'X-User-Role': (window.SmartDisplay.store.getState().authState || {}).role || 'guest'
                }
            })
            .then(function(envelope) {
                console.log('[AlarmControlView] ✅ Disarm response:', envelope);
                var response = envelope.response || {};
                
                if (!response.ok) {
                    throw new Error(response.error || 'DISARM işlemi başarısız');
                }
                
                // Update store with new mode
                window.SmartDisplay.store.setState({
                    alarmoControl: { currentMode: 'disarmed' }
                });
                
                self._showMessage('✓ Alarm başarıyla devre dışı bırakıldı', 'success');
                
                // Immediate refresh + delayed form clear
                self.refresh();
                setTimeout(function() {
                    self._clearForm();
                    self.refresh();
                }, 2000);
            })
            .catch(function(err) {
                console.error('[AlarmControlView] ❌ Disarm error:', err);
                self._showMessage('✗ Hata: ' + (err.message || err), 'error');
            });
        },

        _showMessage: function(text, type) {
            var msgEl = document.getElementById('alarm-message');
            if (!msgEl) return;
            
            msgEl.textContent = text;
            msgEl.className = 'alarm-message ' + type + ' show';
        },

        _clearMessage: function() {
            var msgEl = document.getElementById('alarm-message');
            if (!msgEl) return;
            
            msgEl.className = 'alarm-message';
            msgEl.textContent = '';
        },

        refresh: function() {
            console.log('[AlarmControlView] Refreshing status from store...');
            var state = window.SmartDisplay.store.getState();
            console.log('[AlarmControlView] Full store state:', state);
            
            var alarmoStateObj = state.alarmoState || {};
            console.log('[AlarmControlView] alarmoState object:', alarmoStateObj);
            
            var status = alarmoStateObj.status || {};
            console.log('[AlarmControlView] status object:', status);
            
            this._updateStatus(status);
        },

        setupPolling: function() {
            var self = this;
            console.log('[AlarmControlView] Setting up polling (15s interval)');
			
            this.pollingInterval = setInterval(function() {
                self.fetchStatus();
                self.fetchSensors();
            }, 15000); // 15 seconds
        },

        stopPolling: function() {
            if (this.pollingInterval) {
                console.log('[AlarmControlView] Stopping polling');
                clearInterval(this.pollingInterval);
                this.pollingInterval = null;
            }
        },

        fetchStatus: function() {
            var self = this;
            
            return window.SmartDisplay.api.client.get('/ui/alarmo/status', {
                headers: {
                    'X-User-Role': (window.SmartDisplay.store.getState().authState || {}).role || 'guest'
                }
            })
            .then(function(envelope) {
                console.log('[AlarmControlView] Status response:', envelope);
                var response = envelope.response || {};
                
                if (!response.ok) {
                    throw new Error(response.error || 'Status fetch failed');
                }
                
                // Update Store with fresh HA-backed state before updating UI
                var data = response.data || {};

                var resolvedState = data.alarmo_state || data.alarmo_raw_state || '';
                if (!resolvedState && data.alarmo_mode === 'armed' && data.alarmo_armed_mode) {
                    resolvedState = 'armed_' + data.alarmo_armed_mode;
                } else if (!resolvedState && data.alarmo_mode) {
                    resolvedState = data.alarmo_mode;
                }

                var delayRemaining = typeof data.delay_remaining === 'number' ? data.delay_remaining : 
                                   (typeof data.delay_remaining === 'string' ? parseInt(data.delay_remaining) || 0 : 0);

                console.log('[AlarmControlView] 🐛 DEBUG fetchStatus - Raw delay_remaining:', data.delay_remaining, 'Processed:', delayRemaining);

                window.SmartDisplay.store.setState({
                    alarmoState: {
                        status: {
                            alarmo_connected: data.alarmo_connected || false,
                            ha_runtime_unreachable: data.ha_runtime_unreachable || false,
                            last_seen_at: data.last_seen_at || null,
                            alarmo_state: resolvedState || null,
                            alarmo_mode: data.alarmo_mode || null,
                            alarmo_armed_mode: data.alarmo_armed_mode || null,
                            alarmo_raw_state: data.alarmo_raw_state || null,
                            alarmo_triggered: data.alarmo_triggered === true,
                            delay_remaining: delayRemaining,
                            delay_type: data.delay_type || '',
                            alarmo_last_changed: data.alarmo_last_changed || null
                        },
                        alarmo_state: resolvedState || null,
                        alarmo_triggered: data.alarmo_triggered === true,
                        delay_remaining: delayRemaining
                    },
                    alarmoControl: resolvedState ? { currentMode: resolvedState } : {}
                });

                self._updateStatus(window.SmartDisplay.store.getState().alarmoState.status);
            })
            .catch(function(err) {
                console.error('[AlarmControlView] Status fetch error:', err);
                var stateEl = document.getElementById('alarm-state-display');
                if (stateEl) {
                    stateEl.textContent = '⚠️ Bağlantı Hatası';
                }
            });
        },

        // Fetch sensors and update store
        fetchSensors: function() {
            var self = this;
            return window.SmartDisplay.api.client.get('/ui/alarmo/sensors', {
                headers: {
                    'X-User-Role': (window.SmartDisplay.store.getState().authState || {}).role || 'guest'
                }
            })
            .then(function(envelope) {
                var response = envelope.response || {};
                if (!response.ok) {
                    throw new Error(response.error || 'Sensors fetch failed');
                }
                var data = response.data || [];
                window.SmartDisplay.store.setState({
                    alarmoState: {
                        sensors: Array.isArray(data) ? data : []
                    }
                });
                self.renderSensors();
            })
            .catch(function(err) {
                console.error('[AlarmControlView] Sensors fetch error:', err);
                self.renderSensors([]);
            });
        },

        _updateStatus: function(status) {
            var stateEl = document.getElementById('alarm-state-display');
            if (!stateEl) return;

            // Get alarm state from Store; prefer alarmoState.status.alarmo_state if available, fallback to alarmoControl.currentMode
            var state = window.SmartDisplay.store.getState();
            var alarmoStateObj = state.alarmoState || {};
            var alarmoStatus = alarmoStateObj.status || {};
            var alarmoState = alarmoStatus.alarmo_state || null;

            var alarmoControl = state.alarmoControl || {};
            var controlMode = alarmoControl.currentMode || 'disarmed';

            var currentMode = alarmoState || controlMode;
            
            console.log('[AlarmControlView] Current mode resolved. alarmo_state:', alarmoState, 'controlMode:', controlMode, 'chosen:', currentMode);
            
            var stateText = '🛡️ Devre Dışı';
            
            // Map mode to UI labels
            switch (currentMode) {
                case 'disarmed':
                    stateText = '🛡️ Devre Dışı';
                    break;
                case 'armed_away':
                    stateText = '🔒 Dışarıda Mod';
                    break;
                case 'armed_home':
                    stateText = '🏠 Evde Mod';
                    break;
                case 'armed_night':
                    stateText = '🌙 Gece Mod';
                    break;
                case 'pending':
                case 'arming':
                    stateText = '⏳ Silahlanıyor...';
                    break;
                case 'triggered':
                    stateText = '🚨 TETIKLENDI';
                    break;
                default:
                    stateText = '🔐 ' + (currentMode || 'Etkin');
            }

            stateEl.textContent = stateText;
            console.log('[AlarmControlView] Status updated - Mode:', currentMode, 'Display:', stateText);

            // Update overlay system based on Store state (if available)
            this._updateOverlaySystem(alarmoStatus);
        },

        // Render compact sensor cards under PIN pad
        renderSensors: function() {
            var grid = document.getElementById('sensor-card-grid');
            if (!grid) return;

            var sensors = (window.SmartDisplay.store.getState().alarmoState || {}).sensors || [];
            if (!Array.isArray(sensors) || sensors.length === 0) {
                grid.innerHTML = '<p class="placeholder">Sensör bulunamadı</p>';
                return;
            }

            var html = sensors.map(function(s) {
                var online = s.available !== false && s.state !== 'unavailable';
                var isOpen = String(s.state).toLowerCase() === 'on';
                var statusDotClass = online ? 'dot-online' : 'dot-offline';
                var ocLabel = isOpen ? 'AÇIK' : 'KAPALI';
                var ocClass = isOpen ? 'state-open' : 'state-closed';
                var battPct = (typeof s.battery_percent === 'number' && s.battery_percent >= 0) ? s.battery_percent : null;
                var battText = battPct !== null ? (battPct + '%') : (s.battery_status || '—');

                return [
                    '<div class="sensor-card">',
                    '  <div class="sensor-header">',
                    '    <span class="sensor-dot ' + statusDotClass + '"></span>',
                    '    <span class="sensor-name">' + (s.name || s.id) + '</span>',
                    '  </div>',
                    '  <div class="sensor-body">',
                    '    <span class="sensor-oc ' + ocClass + '">' + ocLabel + '</span>',
                    '    <span class="sensor-battery">Pil: ' + battText + '</span>',
                    '  </div>',
                    '</div>'
                ].join('');
            }).join('');

            grid.innerHTML = html;
        },

        /**
         * Yeni Alarm Overlay Sistemi
         * Kullanıcı özellikleri: 
         * - Alarm kurulduğunda (evde/dışarıda/gece) countdown başlar
         * - Blur arka plan + orta da büyük timer
         * - Son 10 saniyede hızlı pulse
         * - Süre bitince pin pad
         * - Tetiklenince kırmızı blur + pulse + sabit pin pad
         */
        _updateOverlaySystem: function(status) {
            if (!status) {
                this._clearNewOverlays();
                return;
            }

            var state = window.SmartDisplay.store.getState();
            var authState = state.authState || {};
            var userRole = authState.role || 'guest';

            var alarmoState = status.alarmo_state || 'disarmed';
            var isTriggered = status.alarmo_triggered === true;
            var delayRemaining = status.delay_remaining || 0;
            
            console.log('[AlarmControlView] ⚙️ Alarm State Check - State:', alarmoState, 'Triggered:', isTriggered, 'Delay:', delayRemaining);

            // 1. 🚨 ALARM TETİKLENDİ - En yüksek öncelik (Kırmızı blur + pulse + pin pad sabit)
            if (isTriggered) {
                this._showEmergencyMode(userRole);
                return;
            }

            // 2. ⏱️ ALARM KURULUYOR (arming/pending) - Countdown başlat
            if ((alarmoState === 'arming' || alarmoState === 'pending') && delayRemaining > 0) {
                this._startNewCountdown(delayRemaining);
                return;
            }

            // 3. 🛡️ NORMAL DURUM - Ama countdown aktifse dokunma!
            if (this.newCountdownTimer) {
                console.log('[AlarmControlView] ⏲️ Countdown aktif - overlay korunuyor');
                return;
            }
            
            // Countdown yoksa overlayleri temizle
            this._clearNewOverlays();
        },

        /**
         * ⏱️ YENİ COUNTDOWN SİSTEMİ
         * Kullanıcı istekleri:
         * - Alarm kurulunca fullscreen blur arka plan
         * - Ortada büyük geri sayım
         * - Son 10 saniyede hızlı pulse
         * - Süre bitince pin pad
         */
        _startNewCountdown: function(totalSeconds) {
            console.log('[AlarmControlView] 🚀 Başlatılıyor: Yeni countdown -', totalSeconds, 'saniye');
            
            // Mevcut timer'ı temizle
            if (this.newCountdownTimer) {
                clearInterval(this.newCountdownTimer);
                this.newCountdownTimer = null;
            }

            // Fullscreen countdown overlay'i oluştur
            var overlay = this._createNewCountdownOverlay();
            if (!overlay) {
                console.error('[AlarmControlView] Countdown overlay oluşturulamadı!');
                return;
            }

            // Overlay'i göster
            overlay.style.display = 'flex';
            setTimeout(function() { overlay.style.opacity = '1'; }, 50);

            // Timer state
            this.countdownStartTime = Date.now();
            this.countdownTotalSeconds = Math.floor(totalSeconds);
            var remainingSeconds = this.countdownTotalSeconds;

            // Timer display'ı güncelle
            var timerDisplay = overlay.querySelector('.new-countdown-timer');
            if (timerDisplay) {
                timerDisplay.textContent = this._formatTime(remainingSeconds);
                
                // Pulse hızını ayarla
                var pulseSpeed = remainingSeconds <= 10 ? '0.8s' : '2s';
                overlay.style.setProperty('--pulse-speed', pulseSpeed);
                
                if (remainingSeconds <= 10) {
                    timerDisplay.classList.add('warning-pulse');
                }
            }

            // Her saniye güncelleme
            var self = this;
            this.newCountdownTimer = setInterval(function() {
                var elapsed = Math.floor((Date.now() - self.countdownStartTime) / 1000);
                remainingSeconds = Math.max(0, self.countdownTotalSeconds - elapsed);
                
                console.log('[AlarmControlView] ⏱️ Timer güncellemesi:', remainingSeconds, 'saniye kaldı');
                
                // Süre bitti - PIN PAD göster
                if (remainingSeconds <= 0) {
                    clearInterval(self.newCountdownTimer);
                    self.newCountdownTimer = null;
                    
                    console.log('[AlarmControlView] ⏰ Countdown bitti! PIN PAD gösteriliyor...');
                    self._showPinPadMode();
                    return;
                }
                
                // Timer display güncelle
                if (timerDisplay) {
                    timerDisplay.textContent = self._formatTime(remainingSeconds);
                    
                    // Son 10 saniye - hızlı pulse
                    if (remainingSeconds <= 10 && !timerDisplay.classList.contains('warning-pulse')) {
                        console.log('[AlarmControlView] ⚠️ Son 10 saniye! Hızlı pulse aktif.');
                        overlay.style.setProperty('--pulse-speed', '0.4s');
                        timerDisplay.classList.add('warning-pulse');
                        
                        // Circle'a da warning pulse ekle
                        var countdownCircle = overlay.querySelector('.countdown-circle');
                        if (countdownCircle) {
                            countdownCircle.classList.add('warning-pulse');
                        }
                    }
                }
            }, 1000);
            
            console.log('[AlarmControlView] ✅ Countdown başlatıldı -', totalSeconds, 'saniye');
        },

        /**
         * 🎨 YENİ COUNTDOWN OVERLAY OLUŞTURUCU
         * - Fullscreen blur arka plan
         * - Ortada büyük timer
         * - Pulse efekti
         */
        _createNewCountdownOverlay: function() {
            var existingOverlay = document.getElementById('new-countdown-overlay');
            if (existingOverlay) {
                return existingOverlay;
            }

            var overlay = document.createElement('div');
            overlay.id = 'new-countdown-overlay';
            overlay.style.cssText = `
                position: fixed;
                top: 0;
                left: 0;
                width: 100vw;
                height: 100vh;
                z-index: 9999;
                display: none;
                opacity: 0;
                transition: opacity 0.5s ease;
                backdrop-filter: blur(25px);
                background: rgba(0, 0, 0, 0.7);
                justify-content: center;
                align-items: center;
                flex-direction: column;
            `;
            
            // Pulse efekti için ::before pseudo-element benzeri
            var pulseLayer = document.createElement('div');
            pulseLayer.style.cssText = `
                position: absolute;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background: radial-gradient(circle at center, 
                    rgba(255, 0, 0, 0.2) 0%, 
                    rgba(255, 0, 0, 0.1) 40%, 
                    transparent 70%);
                animation: red-blur-pulse var(--pulse-speed, 2s) ease-in-out infinite;
                pointer-events: none;
            `;
            
            // Ana content container
            var content = document.createElement('div');
            content.style.cssText = `
                position: relative;
                z-index: 2;
                text-align: center;
                color: white;
            `;
            
            content.innerHTML = `
                <div style="font-size: 24px; font-weight: 600; margin-bottom: 48px; text-transform: uppercase; letter-spacing: 3px; color: #ffcccc;">
                    Alarm Devreye Alınıyor
                </div>
                <div class="countdown-circle" style="
                    width: 300px;
                    height: 300px;
                    border-radius: 50%;
                    background: linear-gradient(135deg, rgba(255, 0, 0, 0.9), rgba(220, 20, 60, 0.9));
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    margin: 0 auto;
                    box-shadow: 
                        0 0 60px rgba(255, 0, 0, 0.6),
                        inset 0 0 30px rgba(255, 255, 255, 0.1);
                    animation: circle-pulse var(--pulse-speed, 2s) ease-in-out infinite;
                ">
                    <div class="new-countdown-timer" style="
                        font-size: 90px;
                        font-weight: 900;
                        font-family: 'Arial', monospace;
                        color: white;
                        text-shadow: 
                            0 0 20px rgba(255, 255, 255, 0.8),
                            0 4px 20px rgba(0, 0, 0, 0.5);
                        letter-spacing: 4px;
                    ">
                        30
                    </div>
                </div>
                <div style="font-size: 18px; font-weight: 500; margin-top: 32px; opacity: 0.9; color: #ffcccc;">
                    saniye kaldı
                </div>
            `;
            
            // CSS animation ekle
            if (!document.getElementById('countdown-animations')) {
                var style = document.createElement('style');
                style.id = 'countdown-animations';
                style.textContent = `
                    @keyframes red-blur-pulse {
                        0%, 100% { 
                            opacity: 0.6; 
                            transform: scale(1); 
                            background: radial-gradient(circle at center, 
                                rgba(255, 0, 0, 0.2) 0%, 
                                rgba(255, 0, 0, 0.1) 40%, 
                                transparent 70%);
                        }
                        50% { 
                            opacity: 1; 
                            transform: scale(1.05); 
                            background: radial-gradient(circle at center, 
                                rgba(255, 0, 0, 0.4) 0%, 
                                rgba(255, 0, 0, 0.2) 40%, 
                                transparent 70%);
                        }
                    }
                    @keyframes circle-pulse {
                        0%, 100% { 
                            transform: scale(1); 
                            box-shadow: 
                                0 0 60px rgba(255, 0, 0, 0.6),
                                inset 0 0 30px rgba(255, 255, 255, 0.1);
                        }
                        50% { 
                            transform: scale(1.08); 
                            box-shadow: 
                                0 0 100px rgba(255, 0, 0, 0.9),
                                inset 0 0 50px rgba(255, 255, 255, 0.2);
                        }
                    }
                    .warning-pulse {
                        animation: warning-fast-pulse 0.4s ease-in-out infinite !important;
                    }
                    @keyframes warning-fast-pulse {
                        0%, 100% { 
                            transform: scale(1); 
                            background: linear-gradient(135deg, rgba(255, 0, 0, 0.9), rgba(220, 20, 60, 0.9)) !important;
                        }
                        50% { 
                            transform: scale(1.15); 
                            background: linear-gradient(135deg, rgba(255, 50, 50, 1), rgba(255, 100, 100, 1)) !important;
                        }
                    }
                `;
                document.head.appendChild(style);
            }
            
            overlay.appendChild(pulseLayer);
            overlay.appendChild(content);
            document.body.appendChild(overlay);
            
            return overlay;
        },

        /**
         * 📟 PIN PAD MODU
         * Countdown bittikten sonra sadece pin pad gösterilir
         * Siyah blur arka plan
         */
        _showPinPadMode: function() {
            console.log('[AlarmControlView] 📟 PIN PAD modu aktifleştiriliyor...');
            
            // Countdown overlay'ı gizle
            var countdownOverlay = document.getElementById('new-countdown-overlay');
            if (countdownOverlay) {
                countdownOverlay.style.opacity = '0';
                setTimeout(function() {
                    countdownOverlay.style.display = 'none';
                }, 500);
            }
            
            // Pin pad overlay'ı oluştur
            var pinPadOverlay = this._createPinPadOverlay();
            if (!pinPadOverlay) {
                console.error('[AlarmControlView] Pin pad overlay oluşturulamadı!');
                return;
            }
            
            // Pin pad'ı göster
            pinPadOverlay.style.display = 'flex';
            setTimeout(function() {
                pinPadOverlay.style.opacity = '1';
            }, 500);
            
            // PIN PAD EVENT HANDLER'LARINI KURAMASININ KURMASINI KUR!
            this._setupPinPadEvents(pinPadOverlay);
            
            console.log('[AlarmControlView] ✅ PIN PAD modu aktif!');
        },

        /**
         * 📟 PIN PAD OVERLAY OLUŞTURUCU
         * Countdown sonrası minimal pin pad
         */
        _createPinPadOverlay: function() {
            var existingOverlay = document.getElementById('new-pinpad-overlay');
            if (existingOverlay) {
                return existingOverlay;
            }

            var overlay = document.createElement('div');
            overlay.id = 'new-pinpad-overlay';
            overlay.style.cssText = `
                position: fixed;
                top: 0;
                left: 0;
                width: 100vw;
                height: 100vh;
                z-index: 9999;
                display: none;
                opacity: 0;
                transition: opacity 0.5s ease;
                backdrop-filter: blur(25px);
                background: rgba(0, 0, 0, 0.92);
                justify-content: center;
                align-items: center;
                flex-direction: column;
            `;
            
            overlay.innerHTML = `
                <div style="text-align: center; color: white;">
                    <div style="font-size: 32px; font-weight: 600; margin-bottom: 40px;">
                        Alarmı Devre Dışı Bırakmak İçin PIN Giriniz
                    </div>
                    
                    <div style="margin-bottom: 40px;">
                        <input type="password" id="pinpad-input" maxlength="4" 
                               style="
                                   width: 300px;
                                   padding: 24px;
                                   font-size: 48px;
                                   text-align: center;
                                   letter-spacing: 20px;
                                   border: 3px solid rgba(255,255,255,0.3);
                                   border-radius: 12px;
                                   background: rgba(255,255,255,0.1);
                                   color: white;
                                   outline: none;
                               " 
                               placeholder="••••"
                               autocomplete="off">
                    </div>
                    
                    <div style="
                        display: grid;
                        grid-template-columns: repeat(3, 100px);
                        gap: 20px;
                        max-width: 340px;
                    ">
                        <button class="pinpad-btn" data-digit="1" style="
                            padding: 24px;
                            font-size: 28px;
                            font-weight: 600;
                            border: none;
                            border-radius: 12px;
                            background: rgba(255,255,255,0.15);
                            color: white;
                            cursor: pointer;
                            transition: all 0.2s ease;
                        ">1</button>
                        <button class="pinpad-btn" data-digit="2" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.15); color: white; cursor: pointer; transition: all 0.2s ease;">2</button>
                        <button class="pinpad-btn" data-digit="3" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.15); color: white; cursor: pointer; transition: all 0.2s ease;">3</button>
                        <button class="pinpad-btn" data-digit="4" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.15); color: white; cursor: pointer; transition: all 0.2s ease;">4</button>
                        <button class="pinpad-btn" data-digit="5" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.15); color: white; cursor: pointer; transition: all 0.2s ease;">5</button>
                        <button class="pinpad-btn" data-digit="6" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.15); color: white; cursor: pointer; transition: all 0.2s ease;">6</button>
                        <button class="pinpad-btn" data-digit="7" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.15); color: white; cursor: pointer; transition: all 0.2s ease;">7</button>
                        <button class="pinpad-btn" data-digit="8" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.15); color: white; cursor: pointer; transition: all 0.2s ease;">8</button>
                        <button class="pinpad-btn" data-digit="9" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.15); color: white; cursor: pointer; transition: all 0.2s ease;">9</button>
                        
                        <button id="pinpad-clear" style="
                            padding: 24px;
                            font-size: 20px;
                            font-weight: 600;
                            border: none;
                            border-radius: 12px;
                            background: rgba(255,0,0,0.3);
                            color: white;
                            cursor: pointer;
                            transition: all 0.2s ease;
                        ">Temizle</button>
                        
                        <button class="pinpad-btn" data-digit="0" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.15); color: white; cursor: pointer; transition: all 0.2s ease;">0</button>
                        
                        <button id="pinpad-submit" style="
                            padding: 24px;
                            font-size: 20px;
                            font-weight: 600;
                            border: none;
                            border-radius: 12px;
                            background: rgba(0,255,0,0.3);
                            color: white;
                            cursor: pointer;
                            transition: all 0.2s ease;
                        ">Devre Dışı</button>
                    </div>
                    
                    <div style="margin-top: 32px; font-size: 16px; opacity: 0.8;">
                        4 haneli PIN kodunuzu giriniz
                    </div>
                </div>
            `;
            
            document.body.appendChild(overlay);
            return overlay;
        },

        /**
         * 🔧 PIN PAD ETKILEŞİMLERİ
         */
        _setupPinPadEvents: function(overlay) {
            var pinInput = overlay.querySelector('#pinpad-input');
            var digitBtns = overlay.querySelectorAll('.pinpad-btn');
            var clearBtn = overlay.querySelector('#pinpad-clear');
            var submitBtn = overlay.querySelector('#pinpad-submit');
            
            var currentPin = '';
            
            function updateDisplay() {
                pinInput.value = '•'.repeat(currentPin.length);
            }
            
            // Rakam butonları
            digitBtns.forEach(function(btn) {
                btn.addEventListener('click', function() {
                    var digit = this.getAttribute('data-digit');
                    if (currentPin.length < 4) {
                        currentPin += digit;
                        updateDisplay();
                        
                        // 4 rakam girildiğinde otomatik gönder
                        if (currentPin.length === 4) {
                            setTimeout(function() {
                                submitPin();
                            }, 300);
                        }
                    }
                });
                
                // Hover efekti
                btn.addEventListener('mouseenter', function() {
                    this.style.background = 'rgba(255,255,255,0.25)';
                    this.style.transform = 'scale(1.05)';
                });
                btn.addEventListener('mouseleave', function() {
                    this.style.background = 'rgba(255,255,255,0.15)';
                    this.style.transform = 'scale(1)';
                });
            });
            
            // Temizle butonu
            if (clearBtn) {
                clearBtn.addEventListener('click', function() {
                    currentPin = '';
                    updateDisplay();
                });
            }
            
            // Gönder butonu
            if (submitBtn) {
                submitBtn.addEventListener('click', submitPin);
            }
            
            function submitPin() {
                if (currentPin.length !== 4) {
                    console.log('[AlarmControlView] PIN eksik - 4 hane gerekli');
                    return;
                }
                
                console.log('[AlarmControlView] 🔐 PIN gönderiliyor:', currentPin);
                
                // Gerçek alarm disarm API çağrısı
                window.SmartDisplay.api.client.post('/ui/alarmo/disarm', {
                    code: currentPin
                }, {
                    headers: {
                        'X-User-Role': (window.SmartDisplay.store.getState().authState || {}).role || 'admin'
                    }
                })
                .then(function(envelope) {
                    console.log('[AlarmControlView] ✅ Disarm API response:', envelope);
                    var response = envelope.response || {};
                    
                    if (response.ok) {
                        console.log('[AlarmControlView] ✅ PIN doğru! Alarm devre dışı bırakıldı!');
                        
                        // Overlay'i kapat
                        overlay.style.opacity = '0';
                        setTimeout(function() {
                            if (overlay && overlay.parentNode) {
                                overlay.parentNode.removeChild(overlay);
                            }
                        }, 500);
                        
                        // Timer'ları temizle
                        if (overlay._parentView && overlay._parentView.newCountdownTimer) {
                            clearInterval(overlay._parentView.newCountdownTimer);
                            overlay._parentView.newCountdownTimer = null;
                        }
                    } else {
                        throw new Error(response.error || 'Disarm failed');
                    }
                })
                .catch(function(err) {
                    console.log('[AlarmControlView] ❌ PIN hatası:', err.message);
                    // Hata efekti
                    pinInput.style.borderColor = '#f44336';
                    pinInput.style.boxShadow = '0 0 20px rgba(244, 67, 54, 0.5)';
                    
                    // Input'u sallamasın
                    pinInput.style.animation = 'shake 0.5s ease-in-out';
                    
                    setTimeout(function() {
                        pinInput.style.borderColor = 'rgba(255,255,255,0.3)';
                        pinInput.style.boxShadow = 'none';
                        pinInput.style.animation = '';
                        currentPin = '';
                        updateDisplay();
                    }, 1500);
                });
            }
        },
        
        /**
         * 🚨 EMERGENCY MODE
         * Alarm tetiklendiğinde kırmızı blur + pulse + sabit pin pad
         */
        _showEmergencyMode: function(userRole) {
            console.log('[AlarmControlView] 🚨 EMERGENCY MODE aktif! Role:', userRole);
            
            // Diğer overlayleri kapat
            this._clearNewOverlays();
            
            // Emergency overlay oluştur
            var emergencyOverlay = this._createEmergencyOverlay(userRole);
            if (!emergencyOverlay) {
                console.error('[AlarmControlView] Emergency overlay oluşturulamadı!');
                return;
            }
            
            // Emergency overlay'i göster
            emergencyOverlay.style.display = 'flex';
            setTimeout(function() {
                emergencyOverlay.style.opacity = '1';
            }, 50);
            
            console.log('[AlarmControlView] ✅ EMERGENCY MODE aktif!');
        },
        
        /**
         * 🚨 EMERGENCY OVERLAY OLUŞTURUCU
         * Kırmızı blur + pulse + sabit pin pad
         */
        _createEmergencyOverlay: function(userRole) {
            var existingOverlay = document.getElementById('new-emergency-overlay');
            if (existingOverlay) {
                return existingOverlay;
            }

            var overlay = document.createElement('div');
            overlay.id = 'new-emergency-overlay';
            overlay.style.cssText = `
                position: fixed;
                top: 0;
                left: 0;
                width: 100vw;
                height: 100vh;
                z-index: 9999;
                display: none;
                opacity: 0;
                transition: opacity 0.5s ease;
                backdrop-filter: blur(25px);
                background: rgba(139, 0, 0, 0.85);
                justify-content: center;
                align-items: center;
                flex-direction: column;
            `;
            
            // Kırmızı pulse efekti
            var pulseLayer = document.createElement('div');
            pulseLayer.style.cssText = `
                position: absolute;
                top: 0;
                left: 0;
                width: 100%;
                height: 100%;
                background: radial-gradient(circle at center, 
                    rgba(211, 47, 47, 0.3) 0%, 
                    rgba(211, 47, 47, 0.1) 50%, 
                    transparent 80%);
                animation: emergency-pulse 1.2s ease-in-out infinite;
                pointer-events: none;
            `;
            
            var content = '';
            if (userRole !== 'guest') {
                content = `
                    <div style="text-align: center; color: white; position: relative; z-index: 2;">
                        <div style="font-size: 48px; font-weight: 900; margin-bottom: 24px; color: #ffcdd2; text-shadow: 0 0 30px rgba(255, 205, 210, 0.8);">
                            ⚠️ ALARM TETİKLENDİ ⚠️
                        </div>
                        
                        <div style="font-size: 24px; font-weight: 600; margin-bottom: 40px; color: #ffebee;">
                            Alarmı Devre Dışı Bırakmak İçin PIN Giriniz
                        </div>
                        
                        <div style="margin-bottom: 40px;">
                            <input type="password" id="emergency-pinpad-input" maxlength="4" 
                                   style="
                                       width: 300px;
                                       padding: 24px;
                                       font-size: 48px;
                                       text-align: center;
                                       letter-spacing: 20px;
                                       border: 3px solid rgba(255,255,255,0.5);
                                       border-radius: 12px;
                                       background: rgba(255,255,255,0.15);
                                       color: white;
                                       outline: none;
                                   " 
                                   placeholder="••••"
                                   autocomplete="off">
                        </div>
                        
                        <div style="
                            display: grid;
                            grid-template-columns: repeat(3, 100px);
                            gap: 20px;
                            max-width: 340px;
                        ">
                            <button class="emergency-btn" data-digit="1" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.2); color: white; cursor: pointer; transition: all 0.2s ease;">1</button>
                            <button class="emergency-btn" data-digit="2" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.2); color: white; cursor: pointer; transition: all 0.2s ease;">2</button>
                            <button class="emergency-btn" data-digit="3" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.2); color: white; cursor: pointer; transition: all 0.2s ease;">3</button>
                            <button class="emergency-btn" data-digit="4" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.2); color: white; cursor: pointer; transition: all 0.2s ease;">4</button>
                            <button class="emergency-btn" data-digit="5" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.2); color: white; cursor: pointer; transition: all 0.2s ease;">5</button>
                            <button class="emergency-btn" data-digit="6" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.2); color: white; cursor: pointer; transition: all 0.2s ease;">6</button>
                            <button class="emergency-btn" data-digit="7" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.2); color: white; cursor: pointer; transition: all 0.2s ease;">7</button>
                            <button class="emergency-btn" data-digit="8" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.2); color: white; cursor: pointer; transition: all 0.2s ease;">8</button>
                            <button class="emergency-btn" data-digit="9" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.2); color: white; cursor: pointer; transition: all 0.2s ease;">9</button>
                            
                            <button id="emergency-clear" style="padding: 24px; font-size: 20px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,0,0,0.5); color: white; cursor: pointer; transition: all 0.2s ease;">Temizle</button>
                            
                            <button class="emergency-btn" data-digit="0" style="padding: 24px; font-size: 28px; font-weight: 600; border: none; border-radius: 12px; background: rgba(255,255,255,0.2); color: white; cursor: pointer; transition: all 0.2s ease;">0</button>
                            
                            <button id="emergency-submit" style="padding: 24px; font-size: 18px; font-weight: 600; border: none; border-radius: 12px; background: rgba(0,255,0,0.4); color: white; cursor: pointer; transition: all 0.2s ease;">DURDUR</button>
                        </div>
                        
                        <div style="margin-top: 32px; font-size: 16px; opacity: 0.9; color: #ffebee;">
                            ACİL DURUM - 4 haneli PIN kodunuzu giriniz
                        </div>
                    </div>
                `;
            } else {
                content = `
                    <div style="text-align: center; color: white; position: relative; z-index: 2;">
                        <div style="font-size: 48px; font-weight: 900; margin-bottom: 24px; color: #ffcdd2; text-shadow: 0 0 30px rgba(255, 205, 210, 0.8);">
                            ⚠️ ALARM TETİKLENDİ ⚠️
                        </div>
                        
                        <div style="font-size: 24px; font-weight: 600; margin-bottom: 40px; color: #ffebee;">
                            Sadece admin erişimine sahip kullanıcılar<br>
                            alarmı devre dışı bırakabilir
                        </div>
                        
                        <div style="font-size: 18px; opacity: 0.8; color: #ffebee;">
                            Lütfen bir admin ile iletişime geçin
                        </div>
                    </div>
                `;
            }
            
            // CSS animasyonları ekle
            if (!document.getElementById('emergency-animations')) {
                var style = document.createElement('style');
                style.id = 'emergency-animations';
                style.textContent = `
                    @keyframes emergency-pulse {
                        0%, 100% { opacity: 0.8; transform: scale(1); }
                        50% { opacity: 1; transform: scale(1.05); }
                    }
                `;
                document.head.appendChild(style);
            }
            
            overlay.appendChild(pulseLayer);
            overlay.innerHTML += content;
            document.body.appendChild(overlay);
            
            // Emergency PIN PAD etkileşimini kur
            if (userRole !== 'guest') {
                this._setupEmergencyPinPadEvents(overlay);
            }
            
            return overlay;
        },

        /**
         * 🧹 TÜM YENİ OVERLAY'LERİ TEMİZLE
         */
        _clearNewOverlays: function() {
            // Yeni countdown overlay'i kaldır
            var newCountdownOverlay = document.getElementById('new-countdown-overlay');
            if (newCountdownOverlay) {
                newCountdownOverlay.style.opacity = '0';
                setTimeout(function() {
                    if (newCountdownOverlay && newCountdownOverlay.parentNode) {
                        newCountdownOverlay.parentNode.removeChild(newCountdownOverlay);
                    }
                }, 500);
            }
            
            // Yeni PIN pad overlay'i kaldır
            var newPinPadOverlay = document.getElementById('new-pinpad-overlay');
            if (newPinPadOverlay) {
                newPinPadOverlay.style.opacity = '0';
                setTimeout(function() {
                    if (newPinPadOverlay && newPinPadOverlay.parentNode) {
                        newPinPadOverlay.parentNode.removeChild(newPinPadOverlay);
                    }
                }, 500);
            }
            
            // Yeni emergency overlay'i kaldır
            var newEmergencyOverlay = document.getElementById('new-emergency-overlay');
            if (newEmergencyOverlay) {
                newEmergencyOverlay.style.opacity = '0';
                setTimeout(function() {
                    if (newEmergencyOverlay && newEmergencyOverlay.parentNode) {
                        newEmergencyOverlay.parentNode.removeChild(newEmergencyOverlay);
                    }
                }, 500);
            }
            
            // Legacy cleanup
            var countdownScene = document.getElementById('alarm-countdown-scene');
            if (countdownScene) {
                countdownScene.classList.remove('active', 'countdown-scene');
            }

            var emergencyScene = document.getElementById('alarm-emergency-scene');
            if (emergencyScene) {
                emergencyScene.classList.remove('active', 'emergency-scene', 'shake');
            }

            var pinPadScene = document.getElementById('alarm-pinpad-scene');
            if (pinPadScene) {
                pinPadScene.classList.remove('active');
            }

            // Clear legacy overlays for compatibility
            var overlay = document.getElementById('alarm-overlay');
            var countdownOverlay = document.getElementById('alarm-countdown-overlay');
            var triggeredOverlay = document.getElementById('alarm-triggered-overlay');

            if (overlay) {
                overlay.classList.remove('active', 'triggered');
            }

            if (countdownOverlay) {
                countdownOverlay.classList.remove('active', 'warning');
            }

            if (triggeredOverlay) {
                triggeredOverlay.classList.remove('active');
            }
        },

        _createCountdownScene: function() {
            var sceneId = 'alarm-countdown-scene';
            var existingScene = document.getElementById(sceneId);
            
            if (existingScene) {
                return existingScene;
            }

            // Create fullscreen countdown scene with blur background
            var scene = document.createElement('div');
            scene.id = sceneId;
            scene.className = 'alarm-scene-overlay';
            scene.innerHTML = [
                '<div class="countdown-content">',
                '  <div class="countdown-label">Alarm Devreye Alınıyor</div>',
                '  <div class="countdown-timer">30</div>',
                '</div>'
            ].join('\n');

            document.body.appendChild(scene);
            return scene;
        },

        _createEmergencyScene: function(userRole) {
            var sceneId = 'alarm-emergency-scene';
            var existingScene = document.getElementById(sceneId);
            
            if (existingScene) {
                return existingScene;
            }

            // Create fullscreen emergency scene with red blur background
            var scene = document.createElement('div');
            scene.id = sceneId;
            scene.className = 'alarm-scene-overlay';
            
            var pinSection = '';
            if (userRole !== 'guest') {
                pinSection = [
                    '<div class="emergency-pin-section">',
                    '  <input type="password" class="emergency-pin-input" maxlength="4" placeholder="PIN" autocomplete="off">',
                    '  <div class="emergency-pin-grid">',
                    '    <button class="emergency-pin-btn" data-digit="1">1</button>',
                    '    <button class="emergency-pin-btn" data-digit="2">2</button>',
                    '    <button class="emergency-pin-btn" data-digit="3">3</button>',
                    '    <button class="emergency-pin-btn" data-digit="4">4</button>',
                    '    <button class="emergency-pin-btn" data-digit="5">5</button>',
                    '    <button class="emergency-pin-btn" data-digit="6">6</button>',
                    '    <button class="emergency-pin-btn" data-digit="7">7</button>',
                    '    <button class="emergency-pin-btn" data-digit="8">8</button>',
                    '    <button class="emergency-pin-btn" data-digit="9">9</button>',
                    '    <button class="emergency-pin-btn emergency-pin-clear">Temizle</button>',
                    '    <button class="emergency-pin-btn" data-digit="0">0</button>',
                    '    <button class="emergency-pin-btn emergency-pin-submit">Devre Dışı</button>',
                    '  </div>',
                    '</div>'
                ].join('\n');
            } else {
                pinSection = '<div class="emergency-guest-warning">Sadece admin erişimine sahip kullanıcılar alarmı devre dışı bırakabilir</div>';
            }

            scene.innerHTML = [
                '<div class="emergency-content">',
                '  <div class="emergency-title">ALARM TETİKLENDİ</div>',
                '  <div class="emergency-subtitle">Devre dışı bırakmak için PIN giriniz</div>',
                '  <div class="emergency-countdown" style="display:none;">',
                '    <div class="emergency-countdown-value">30</div>',
                '  </div>',
                pinSection,
                '</div>'
            ].join('\n');

            document.body.appendChild(scene);
            return scene;
        },

        _showPinPadScene: function() {
            console.log('[AlarmControlView] 📟 Showing PIN pad scene after countdown');
            
            // Hide countdown scene
            var countdownScene = document.getElementById('alarm-countdown-scene');
            if (countdownScene) {
                countdownScene.classList.remove('active');
            }

            // Get current user role
            var state = window.SmartDisplay.store.getState();
            var authState = state.authState || {};
            var userRole = authState.role || 'guest';

            // Create and show pin pad scene (minimal design)
            var pinPadScene = this._createPinPadScene(userRole);
            if (pinPadScene) {
                pinPadScene.classList.add('active');
                this._setupPinPadInteraction(pinPadScene);
            }
        },

        _createPinPadScene: function(userRole) {
            var sceneId = 'alarm-pinpad-scene';
            var existingScene = document.getElementById(sceneId);
            
            if (existingScene) {
                return existingScene;
            }

            // Create minimal fullscreen pin pad scene
            var scene = document.createElement('div');
            scene.id = sceneId;
            scene.className = 'alarm-scene-overlay';
            scene.style.cssText = 'background: rgba(0, 0, 0, 0.95); backdrop-filter: blur(20px);';
            
            var content = '';
            if (userRole !== 'guest') {
                content = [
                    '<div style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%; gap: 32px;">',
                    '  <div style="color: white; font-size: 24px; text-align: center;">Alarmı devre dışı bırakmak için PIN giriniz</div>',
                    '  <input type="password" style="width: 300px; padding: 20px; font-size: 32px; text-align: center; letter-spacing: 10px; border: 2px solid #666; border-radius: 8px; background: rgba(255,255,255,0.1); color: white;" maxlength="4" placeholder="••••" autocomplete="off" id="final-pin-input">',
                    '  <div style="display: grid; grid-template-columns: repeat(3, 80px); gap: 16px; max-width: 280px;">',
                    '    <button style="padding: 20px; font-size: 24px; border: none; border-radius: 8px; background: rgba(255,255,255,0.1); color: white; cursor: pointer;" data-digit="1">1</button>',
                    '    <button style="padding: 20px; font-size: 24px; border: none; border-radius: 8px; background: rgba(255,255,255,0.1); color: white; cursor: pointer;" data-digit="2">2</button>',
                    '    <button style="padding: 20px; font-size: 24px; border: none; border-radius: 8px; background: rgba(255,255,255,0.1); color: white; cursor: pointer;" data-digit="3">3</button>',
                    '    <button style="padding: 20px; font-size: 24px; border: none; border-radius: 8px; background: rgba(255,255,255,0.1); color: white; cursor: pointer;" data-digit="4">4</button>',
                    '    <button style="padding: 20px; font-size: 24px; border: none; border-radius: 8px; background: rgba(255,255,255,0.1); color: white; cursor: pointer;" data-digit="5">5</button>',
                    '    <button style="padding: 20px; font-size: 24px; border: none; border-radius: 8px; background: rgba(255,255,255,0.1); color: white; cursor: pointer;" data-digit="6">6</button>',
                    '    <button style="padding: 20px; font-size: 24px; border: none; border-radius: 8px; background: rgba(255,255,255,0.1); color: white; cursor: pointer;" data-digit="7">7</button>',
                    '    <button style="padding: 20px; font-size: 24px; border: none; border-radius: 8px; background: rgba(255,255,255,0.1); color: white; cursor: pointer;" data-digit="8">8</button>',
                    '    <button style="padding: 20px; font-size: 24px; border: none; border-radius: 8px; background: rgba(255,255,255,0.1); color: white; cursor: pointer;" data-digit="9">9</button>',
                    '    <button style="padding: 20px; font-size: 18px; border: none; border-radius: 8px; background: rgba(255,0,0,0.2); color: white; cursor: pointer; grid-column: span 1;" id="pin-clear">C</button>',
                    '    <button style="padding: 20px; font-size: 24px; border: none; border-radius: 8px; background: rgba(255,255,255,0.1); color: white; cursor: pointer;" data-digit="0">0</button>',
                    '    <button style="padding: 20px; font-size: 18px; border: none; border-radius: 8px; background: rgba(0,255,0,0.2); color: white; cursor: pointer; grid-column: span 1;" id="pin-submit">OK</button>',
                    '  </div>',
                    '</div>'
                ].join('\n');
            } else {
                content = '<div style="display: flex; align-items: center; justify-content: center; height: 100%; color: white; font-size: 24px; text-align: center;">Sadece admin erişimine sahip kullanıcılar alarmı devre dışı bırakabilir</div>';
            }

            scene.innerHTML = content;
            document.body.appendChild(scene);
            return scene;
        },

        _setupPinPadInteraction: function(pinPadScene) {
            var pinInput = pinPadScene.querySelector('#final-pin-input');
            var digitBtns = pinPadScene.querySelectorAll('[data-digit]');
            var clearBtn = pinPadScene.querySelector('#pin-clear');
            var submitBtn = pinPadScene.querySelector('#pin-submit');

            if (!pinInput) return;

            // Pin input
            var currentPin = '';
            
            function updatePinDisplay() {
                pinInput.value = '•'.repeat(currentPin.length);
            }

            // Digit buttons
            digitBtns.forEach(function(btn) {
                btn.addEventListener('click', function() {
                    var digit = this.getAttribute('data-digit');
                    if (currentPin.length < 4) {
                        currentPin += digit;
                        updatePinDisplay();
                        
                        // Auto submit when 4 digits entered
                        if (currentPin.length === 4) {
                            setTimeout(function() {
                                submitPin();
                            }, 500);
                        }
                    }
                });
            });

            // Clear button
            if (clearBtn) {
                clearBtn.addEventListener('click', function() {
                    currentPin = '';
                    updatePinDisplay();
                });
            }

            // Submit button
            if (submitBtn) {
                submitBtn.addEventListener('click', submitPin);
            }

            function submitPin() {
                if (currentPin.length !== 4) return;

                console.log('[AlarmControlView] Attempting to disarm with PIN');
                
                // Call disarm API
                window.SmartDisplay.api.disarmAlarm(currentPin)
                    .then(function(response) {
                        console.log('[AlarmControlView] Alarm disarmed successfully');
                        pinPadScene.classList.remove('active');
                    })
                    .catch(function(error) {
                        console.error('[AlarmControlView] Failed to disarm:', error);
                        // Show error and reset PIN
                        currentPin = '';
                        updatePinDisplay();
                        pinInput.style.borderColor = '#f44336';
                        setTimeout(function() {
                            pinInput.style.borderColor = '#666';
                        }, 1000);
                    });
            }
        },

        _setupEmergencyPinPad: function(emergencyScene) {
            var pinInput = emergencyScene.querySelector('.emergency-pin-input');
            var digitBtns = emergencyScene.querySelectorAll('[data-digit]');
            var clearBtn = emergencyScene.querySelector('.emergency-pin-clear');
            var submitBtn = emergencyScene.querySelector('.emergency-pin-submit');

            if (!pinInput) return;

            var currentPin = '';
            
            function updatePinDisplay() {
                pinInput.value = '•'.repeat(currentPin.length);
            }

            // Digit buttons
            digitBtns.forEach(function(btn) {
                btn.addEventListener('click', function() {
                    var digit = this.getAttribute('data-digit');
                    if (currentPin.length < 4) {
                        currentPin += digit;
                        updatePinDisplay();
                    }
                });
            });

            // Clear button
            if (clearBtn) {
                clearBtn.addEventListener('click', function() {
                    currentPin = '';
                    updatePinDisplay();
                });
            }

            // Submit button
            if (submitBtn) {
                submitBtn.addEventListener('click', function() {
                    if (currentPin.length === 4) {
                        console.log('[AlarmControlView] Emergency disarm attempt');
                        
                        window.SmartDisplay.api.disarmAlarm(currentPin)
                            .then(function(response) {
                                console.log('[AlarmControlView] Emergency alarm disarmed');
                                emergencyScene.classList.remove('active');
                            })
                            .catch(function(error) {
                                console.error('[AlarmControlView] Emergency disarm failed:', error);
                                currentPin = '';
                                updatePinDisplay();
                                pinInput.style.borderColor = '#f44336';
                                setTimeout(function() {
                                    pinInput.style.borderColor = 'rgba(211, 47, 47, 0.5)';
                                }, 1000);
                            });
                    }
                });
            }
        },

        /**
         * 🔄 LEGACY COMPATIBILITY FUNCTIONS
         * Eski fonksiyonları yeni sisteme yönlendiriyor
         */
        _clearAllOverlays: function() {
            // Legacy compatibility - redirect to new system
            this._clearNewOverlays();
        },

        _formatCountdown: function(seconds) {
            // Legacy compatibility - redirect to new system 
            return this._formatTime(seconds);
        },

        /**
         * ⏰ TIME FORMATTER
         * Sadece saniye sayısını döndürür (yuvarlak tasarım için)
         */
        _formatTime: function(seconds) {
            return seconds.toString();
        },

        _showPinPadScene: function() {
            // Legacy compatibility - redirect to new system
            this._showPinPadMode();
        },

        _createPinPadScene: function(userRole) {
            // Legacy compatibility - redirect to new system
            return this._createPinPadOverlay();
        },

        _setupPinPadInteraction: function(pinPadScene) {
            // Legacy compatibility - redirect to new system
            this._setupPinPadEvents(pinPadScene);
        },

        _setupEmergencyPinPad: function(emergencyScene) {
            // Legacy compatibility - redirect to new system
            this._setupEmergencyPinPadEvents(emergencyScene);
        },

        // DEBUG: Manual countdown trigger for testing
        testCountdown: function(seconds) {
            console.log('[AlarmControlView] 🧪 Testing countdown overlay with', seconds, 'seconds');
            this._startNewCountdown(seconds || 30);
        },

        // Manual countdown test with Turkish user's exact specs: 30 sec with fast pulse at last 10
        testCountdownTurkishSpecs: function() {
            console.log('[AlarmControlView] 🇹🇷 Testing Turkish user countdown specs...');
            this._startNewCountdown(30); // Direct call to new countdown system
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
            
            // Check if guest mode is active (FAZ L2)
            var state = window.SmartDisplay.store.getState();
            var authState = state.authState || {};
            var guestState = state.guestState || {};
            var isGuest = guestState.active && authState.role === 'guest';
            
            var html = [
                '<div class="menu-split-layout">',
                '  <div class="menu-sidebar">',
                '    <div class="menu-container">',
                '      <div class="menu-header">',
                '        <h2 class="menu-title">Menu</h2>'
            ];
            
            // Add guest badge if in guest mode
            if (isGuest) {
                html.push('        <span class="guest-badge">Guest Access</span>');
            }
            
            html = html.concat([
                '      </div>',
                '      <div class="menu-content" id="menu-content">',
                '        <!-- Menu sections rendered here -->',
                '      </div>',
                '      <div class="menu-error" id="menu-error" style="display:none;"></div>',
                '      <div class="menu-footer">'
            ]);
            
            // Add end guest session button if in guest mode
            if (isGuest) {
                html.push('        <button class="menu-end-guest-btn" id="menu-end-guest-btn">End Guest Session</button>');
            }
            
            html = html.concat([
                '        <button class="menu-close-btn" id="menu-close-btn">Close</button>',
                '      </div>',
                '    </div>',
                '  </div>',
                '  <div class="menu-content-area" id="menu-content-area">',
                '    <!-- Content view rendered here -->',
                '  </div>',
                '</div>'
            ]);
            
            viewElement.innerHTML = html.join('\n');
            
            overlay.appendChild(viewElement);
            
            // Setup event listeners
            this._setupEventListeners();
            
            // Schedule menu auto-hide timer (10s idle, 5m full close)
            if (window.SmartDisplay.viewManager) {
                var homeView = window.SmartDisplay.viewManager.views['home'];
                if (homeView && homeView._scheduleMenuAutoHide) {
                    console.log('[MenuView] Scheduling menu auto-hide timer on mount');
                    homeView._scheduleMenuAutoHide();
                }
            }
            
            // Initialize controller and render menu
            var self = this;
            this._initController();
            
            // Auto-select Alarm view after menu initialization
            setTimeout(function() {
                self._selectAlarmByDefault();
            }, 100);
            
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
            
            // Re-render menu if role changed (e.g., after login)
            var currentState = window.SmartDisplay.store.getState();
            var currentRole = currentState.currentRole || 'guest';
            if (!this.lastRole) {
                this.lastRole = currentRole;
            }
            if (currentRole !== this.lastRole) {
                console.log('[MenuView] Role changed from ' + this.lastRole + ' to ' + currentRole + ', re-rendering menu');
                this.lastRole = currentRole;
                this._renderMenu();
            }
            
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

            // End guest session button (FAZ L2)
            var endGuestBtn = document.getElementById('menu-end-guest-btn');
            if (endGuestBtn) {
                endGuestBtn.addEventListener('click', function() {
                    self._handleEndGuestSession();
                });
            }

            // Click outside menu to close (on overlay background)
            var overlay = document.getElementById('menu-overlay');
            if (overlay) {
                overlay.addEventListener('click', function(e) {
                    console.log('[MenuView] Overlay click target:', e.target && e.target.className);
                    // Only close if clicking on overlay background, not the menu itself
                    if (e.target === overlay) {
                        self._handleClose();
                    }
                    
                    // Menu item clicks via event delegation on overlay
                    // Check if clicked element or its parent is an expandable header
                    var expandableHeader = e.target.closest('.menu-section-expandable');
                    if (expandableHeader) {
                        console.log('[MenuView] Expandable header clicked:', expandableHeader.textContent);
                        window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                        e.stopPropagation();
                        e.preventDefault();
                        self._toggleSubsections(expandableHeader);
                        return;
                    }
                    
                    // Check if clicked element is a menu item (but not a header button)
                    var menuItem = e.target.closest('.menu-item');
                    var headerBtn = e.target.closest('.menu-section-header');
                    
                    if (menuItem && !headerBtn) {
                        console.log('[MenuView] Menu item clicked:', menuItem.getAttribute('data-section'));
                        window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                        self._handleMenuItemClick(menuItem);
                    }
                });
            }

            // Container-level delegation (backup in case overlay listener misses)
            container.addEventListener('click', function(e) {
                // Expandable headers take priority - they toggle subsections, don't navigate
                var expandableHeader = e.target.closest('.menu-section-expandable');
                if (expandableHeader) {
                    console.log('[MenuView] (container) Expandable header clicked:', expandableHeader.textContent);
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
                    e.stopPropagation();
                    e.preventDefault();
                    self._toggleSubsections(expandableHeader);
                    return;
                }

                // Any section header (expandable or not) shouldn't navigate - it's just a label
                var headerBtn = e.target.closest('.menu-section-header');
                if (headerBtn) {
                    console.log('[MenuView] (container) Header button clicked (non-expandable):', headerBtn.textContent);
                    // Non-expandable headers don't navigate either
                    return;
                }

                // Menu items (actual action buttons) navigate
                var menuItem = e.target.closest('.menu-item');
                if (menuItem && !menuItem.disabled) {
                    console.log('[MenuView] (container) Menu item clicked:', menuItem.getAttribute('data-action') || menuItem.getAttribute('data-subsection') || menuItem.getAttribute('data-section'));
                    window.SmartDisplay.viewManager.resetMenuAutoHideTimer();
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
            var currentRole = window.SmartDisplay.store.getState().currentRole || 'guest';

            if (!sections || sections.length === 0) {
                console.warn('[MenuView] No visible sections');
                this._renderFallback();
                return;
            }

            var contentEl = document.getElementById('menu-content');
            if (!contentEl) return;

            contentEl.innerHTML = '';
            var self = this;

            // Render each section
            sections.forEach(function(section) {
                var sectionEl = document.createElement('div');
                sectionEl.className = 'menu-section';
                sectionEl.setAttribute('data-section-id', section.id);

                // Section header - Make it expandable if has sub_sections
                if (section.name) {
                    var headerEl = document.createElement('button');
                    headerEl.className = 'menu-section-header menu-section-button';
                    headerEl.textContent = section.name;
                    headerEl.setAttribute('data-section', section.id);
                    
                    // Add expand class if has subsections (CSS will add icon)
                    if (section.sub_sections && section.sub_sections.length > 0) {
                        headerEl.classList.add('menu-section-expandable');
                    }
                    
                    sectionEl.appendChild(headerEl);
                }

                // If section has subsections, render both actions and subsections inside the collapsible container
                if (section.sub_sections && section.sub_sections.length > 0) {
                    var subContainerEl = document.createElement('div');
                    subContainerEl.className = 'menu-subsections';
                    subContainerEl.setAttribute('data-parent-section', section.id);

                    // Main actions (hidden until expanded)
                    if (section.actions && Array.isArray(section.actions)) {
                        section.actions.forEach(function(item) {
                            if (item.enabled === false) {
                                return;
                            }

                            var itemEl = document.createElement('button');
                            itemEl.className = 'menu-item';
                            itemEl.setAttribute('data-section', section.id);
                            itemEl.setAttribute('data-action', item.id);

                            var labelEl = document.createElement('span');
                            labelEl.className = 'menu-item-label';
                            labelEl.textContent = item.name || item.id;
                            itemEl.appendChild(labelEl);

                            subContainerEl.appendChild(itemEl);
                        });
                    }

                    // Subsections
                    section.sub_sections.forEach(function(subsection) {
                        if (subsection.visible === false) {
                            return;
                        }

                        var subEl = document.createElement('button');
                        subEl.className = 'menu-item menu-subitem';
                        subEl.setAttribute('data-section', section.id);
                        subEl.setAttribute('data-subsection', subsection.id);

                        var subLabelEl = document.createElement('span');
                        subLabelEl.className = 'menu-item-label';
                        subLabelEl.textContent = subsection.name || subsection.id;
                        subEl.appendChild(subLabelEl);

                        subContainerEl.appendChild(subEl);
                    });

                    sectionEl.appendChild(subContainerEl);
                } else if (section.actions && Array.isArray(section.actions)) {
                    // No subsections: render actions normally (visible by default)
                    var itemsContainerEl = document.createElement('div');
                    itemsContainerEl.className = 'menu-items';

                    section.actions.forEach(function(item) {
                        if (item.enabled === false) {
                            return;
                        }

                        var itemEl = document.createElement('button');
                        itemEl.className = 'menu-item';
                        itemEl.setAttribute('data-section', section.id);
                        itemEl.setAttribute('data-action', item.id);

                        var labelEl = document.createElement('span');
                        labelEl.className = 'menu-item-label';
                        labelEl.textContent = item.name || item.id;
                        itemEl.appendChild(labelEl);

                        itemsContainerEl.appendChild(itemEl);
                    });

                    sectionEl.appendChild(itemsContainerEl);
                }

                contentEl.appendChild(sectionEl);
            });

            // Update highlights
            this._updateHighlight();

            console.log('[MenuView] Render complete');
            
            // Debug: show expandable headers found
            var expandableHeaders = document.querySelectorAll('.menu-section-expandable');
            console.log('[MenuView] Found ' + expandableHeaders.length + ' expandable headers');
        },

        _toggleSubsections: function(headerBtn) {
            var sectionEl = headerBtn.closest('.menu-section');
            if (!sectionEl) return;

            var subsectionsEl = sectionEl.querySelector('.menu-subsections');
            if (!subsectionsEl) return;

            var isOpen = subsectionsEl.classList.contains('menu-subsections-open');
            console.log('[MenuView] Toggling subsections, currently ' + (isOpen ? 'open' : 'closed'));

            if (isOpen) {
                subsectionsEl.classList.remove('menu-subsections-open');
                headerBtn.classList.remove('menu-section-expanded');
            } else {
                subsectionsEl.classList.add('menu-subsections-open');
                headerBtn.classList.add('menu-section-expanded');
            }
        },

        _renderFallback: function() {
            var contentEl = document.getElementById('menu-content');
            if (!contentEl) return;

            contentEl.innerHTML = [
                '<div class="menu-section">',
                '  <div class="menu-items">',
                '    <button class="menu-item" data-view="home">Home</button>',
                '    <button class="menu-item" data-view="alarm">Alarm</button>',
                '    <button class="menu-item" data-view="alarm-control">Alarm Kontrolü</button>',
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
            // Prefer the most specific target: action id, then subsection id, then section/view id.
            var viewId = btn.getAttribute('data-action')
                || btn.getAttribute('data-subsection')
                || btn.getAttribute('data-section')
                || btn.getAttribute('data-view');

            if (!viewId) {
                console.warn('[MenuView] Menu item click without viewId', btn);
                return;
            }

            console.log('[MenuView] Navigating to:', viewId, 'Button:', btn.className);

            if (window.SmartDisplay.viewManager && window.SmartDisplay.viewManager.isAlarmLocked()) {
                console.log('[MenuView] Alarm lock active, ignoring menu navigation');
                return;
            }

            console.log('[MenuView] Menu item clicked:', viewId);

            // Special handling for alarm_control - show in content area and keep menu open
            if (viewId === 'alarm_control') {
                console.log('[MenuView] Showing alarm control in content area');
                
                // Update HomeView's savedMenuView to remember alarm is active
                var homeView = window.SmartDisplay.viewManager.views['home'];
                if (homeView) {
                    homeView.savedMenuView = 'alarm_control';
                    console.log('[MenuView] Updated savedMenuView to alarm_control');
                }
                
                // Update store to track that alarm_control is open
                if (window.SmartDisplay.store) {
                    window.SmartDisplay.store.setState({
                        menu: {
                            currentView: viewId
                        }
                    });
                }
                
                // Show alarm in the content area
                this._showAlarmContent();
                
                // Keep menu open - don't close
                return;
            }

            // Update store and trigger navigation
            if (window.SmartDisplay.store) {
                window.SmartDisplay.store.setState({
                    menu: {
                        currentView: viewId
                    }
                });
            }

            // Close menu after selection
            if (window.SmartDisplay.viewManager) {
                window.SmartDisplay.viewManager.closeMenu();
            }
        },

        _showAlarmContent: function() {
            var contentArea = document.getElementById('menu-content-area');
            if (!contentArea) {
                console.error('[MenuView] Content area not found');
                return;
            }

            // Clear previous content
            contentArea.innerHTML = '';

            // Mount AlarmControlView in content area
            var alarmView = window.SmartDisplay.viewManager.views['alarm-control'];
            if (alarmView) {
                // Create wrapper div for the alarm view
                var wrapper = document.createElement('div');
                wrapper.id = 'alarm-content-wrapper';
                wrapper.style.width = '100%';
                wrapper.style.height = '100%';
                wrapper.style.display = 'flex';
                wrapper.style.flexDirection = 'column';
                wrapper.style.overflow = 'hidden';

                contentArea.appendChild(wrapper);

                // Mount the alarm view
                if (alarmView.mount && typeof alarmView.mount === 'function') {
                    // Temporarily change the mount target
                    var originalMount = alarmView.mount;
                    alarmView.mount = function() {
                        console.log('[AlarmControlView] Mounting in content area (modern glass glow)');
                        var mainEl = wrapper;
                        if (mainEl) {
                            mainEl.innerHTML = '';
                        }

                        // Render modern glass glow alarm HTML
                        mainEl.innerHTML = `
                            <!-- Alarm Overlay System -->
                            <div class="alarm-overlay" id="alarm-overlay"></div>
                            
                            <!-- Countdown Overlay (pending/arming) -->
                            <div class="alarm-countdown-overlay" id="alarm-countdown-overlay">
                                <div class="alarm-countdown-label">Alarm will activate in</div>
                                <div class="alarm-countdown-value" id="alarm-countdown-value">--</div>
                            </div>
                            
                            <!-- Triggered State PIN Overlay -->
                            <div class="alarm-triggered-overlay" id="alarm-triggered-overlay">
                                <div class="triggered-header">
                                    <div class="triggered-title">⚠️ ALARM TRIGGERED</div>
                                    <div class="triggered-subtitle">Immediate action required</div>
                                </div>
                                
                                <div class="triggered-pin-section" id="triggered-pin-section">
                                    <div class="triggered-pin-label">Enter PIN to disarm</div>
                                    <input type="password" id="triggered-pin-input" class="triggered-pin-input" maxlength="6" placeholder="••••••" autocomplete="off">
                                    
                                    <div class="triggered-numpad">
                                        <button class="triggered-numpad-btn" data-num="1">1</button>
                                        <button class="triggered-numpad-btn" data-num="2">2</button>
                                        <button class="triggered-numpad-btn" data-num="3">3</button>
                                        <button class="triggered-numpad-btn" data-num="4">4</button>
                                        <button class="triggered-numpad-btn" data-num="5">5</button>
                                        <button class="triggered-numpad-btn" data-num="6">6</button>
                                        <button class="triggered-numpad-btn" data-num="7">7</button>
                                        <button class="triggered-numpad-btn" data-num="8">8</button>
                                        <button class="triggered-numpad-btn" data-num="9">9</button>
                                        <button class="triggered-numpad-btn" data-num="0">0</button>
                                        <button class="triggered-numpad-btn clear" data-num="clear">⌫</button>
                                    </div>
                                    
                                    <button class="triggered-disarm-btn" id="triggered-disarm-btn">Disarm Now</button>
                                </div>
                                
                                <div class="triggered-guest-warning" id="triggered-guest-warning" style="display: none;">
                                    <p>🔒 Guest mode active<br>Only administrators can disarm the alarm</p>
                                </div>
                                
                                <div class="triggered-countdown" id="triggered-countdown" style="display: none;">
                                    <div class="triggered-countdown-label">Alarm escalation in</div>
                                    <div class="triggered-countdown-value" id="triggered-countdown-value">--</div>
                                </div>
                            </div>
                            
                            <div class="alarm-container">
                                <!-- Durum Göstergesi -->
                                <div class="alarm-state-indicator">
                                    <h3>Mevcut Durum</h3>
                                    <div class="state-value" id="alarm-state-display">Yükleniyor...</div>
                                </div>

                                <!-- Mesaj Gösterimi -->
                                <div id="alarm-message" class="alarm-message"></div>

                                <!-- Mod Seçim Paneli -->
                                <div class="alarm-modes-section">
                                    <label class="alarm-modes-label">Güvenlik Modu Seçiniz</label>
                                    <div class="alarm-modes">
                                        <button class="alarm-mode-btn" data-mode="disarmed">
                                            <div class="mode-icon">🛡️</div>
                                            <div class="mode-label">Devre Dışı</div>
                                        </button>
                                        <button class="alarm-mode-btn" data-mode="armed_away">
                                            <div class="mode-icon">🔒</div>
                                            <div class="mode-label">Dışarıda</div>
                                        </button>
                                        <button class="alarm-mode-btn" data-mode="armed_home">
                                            <div class="mode-icon">🏠</div>
                                            <div class="mode-label">Evde</div>
                                        </button>
                                        <button class="alarm-mode-btn" data-mode="armed_night">
                                            <div class="mode-icon">🌙</div>
                                            <div class="mode-label">Gece</div>
                                        </button>
                                    </div>
                                </div>

                                <!-- Kimlik Doğrulama Paneli -->
                                <div class="alarm-code-section">
                                    <label>Güvenlik Kodu</label>
                                    <input type="password" id="alarm-code-input" class="alarm-code-input" maxlength="6" placeholder="••••••" autocomplete="off">
                                    
                                    <!-- Sayısal Tuş Takımı -->
                                    <div class="alarm-numpad">
                                        <button class="numpad-btn" data-num="1">1</button>
                                        <button class="numpad-btn" data-num="2">2</button>
                                        <button class="numpad-btn" data-num="3">3</button>
                                        
                                        <button class="numpad-btn" data-num="4">4</button>
                                        <button class="numpad-btn" data-num="5">5</button>
                                        <button class="numpad-btn" data-num="6">6</button>
                                        
                                        <button class="numpad-btn" data-num="7">7</button>
                                        <button class="numpad-btn" data-num="8">8</button>
                                        <button class="numpad-btn" data-num="9">9</button>
                                        
                                        <button class="numpad-btn" data-num="0">0</button>
                                        <button class="numpad-btn clear" data-num="clear">⌫</button>
                                    </div>
                                </div>

                                <!-- İşlem Butonları -->
                            </div>
                        `;

                        // Setup event listeners
                        alarmView._setupEventListeners();
                        alarmView._setupTriggeredOverlayListeners();
                        
                        // Initial status fetch and polling setup
                        alarmView.refresh();
                        alarmView.setupPolling();

                        // Event listeners are set up by AlarmControlView._setupEventListeners
                        // Activity detection is not needed here as it interferes with button handlers
                    };

                    alarmView.mount();
                }
            }
        },

        _handleClose: function() {
            console.log('[MenuView] Close button clicked');
            // Close the menu overlay
            if (window.SmartDisplay.viewManager) {
                window.SmartDisplay.viewManager.closeMenu();
            }
        },

        _handleEndGuestSession: function() {
            console.log('[MenuView] End guest session clicked');
            
            // Confirm action
            if (!confirm('End guest session? You will be logged out.')) {
                return;
            }
            
            // Clear auth and guest state (FAZ L2)
            if (window.SmartDisplay.store) {
                window.SmartDisplay.store.setState({
                    authState: {
                        authenticated: false,
                        role: 'guest',
                        pin: null
                    },
                    guestState: {
                        active: false,
                        requestId: null,
                        targetUser: null,
                        approvalTime: null,
                        pollingActive: false
                    }
                });
            }
            
            // Close menu and route to login
            if (window.SmartDisplay.viewManager) {
                window.SmartDisplay.viewManager.closeMenu();
                setTimeout(function() {
                    window.SmartDisplay.viewManager.routeToView('login');
                }, 300);
            }
        },

        _showError: function(message) {
            var errorEl = document.getElementById('menu-error');
            if (errorEl) {
                errorEl.textContent = message;
                errorEl.style.display = 'block';
            }
        },

        _selectAlarmByDefault: function() {
            console.log('[MenuView] Auto-selecting alarm control on menu open');
            
            // Update HomeView's savedMenuView to alarm_control since we're showing alarm
            var homeView = window.SmartDisplay.viewManager.views['home'];
            if (homeView) {
                homeView.savedMenuView = 'alarm_control';
                console.log('[MenuView] Set savedMenuView to alarm_control');
            }
            
            // Find the alarm menu item (data-action="alarm_control")
            var alarmItem = document.querySelector('[data-action="alarm_control"]');
            if (alarmItem) {
                console.log('[MenuView] Found alarm item, clicking it');
                alarmItem.click();
            }
        }
    };

    // ========================================================================
    // View Manager
    // ========================================================================
    var ViewManager = {
        currentView: null,
        views: {
            'login': LoginView,
            'guest-request': GuestRequestView,
            'first-boot': FirstBootView,
            'home': HomeView,
            'alarm': AlarmView,
            'guest': GuestView,
            'settings': SettingsView,
            'alarmo-settings': AlarmoSettingsView,
            'alarm-control': AlarmControlView,
            'menu': MenuView
        },

        // ====================================================================
        // View Switching
        // ====================================================================

        /**
         * Route to a specific view by ID
         * FAZ L1: Helper for programmatic routing (e.g. after login)
         * @param {string} viewId - View to route to
         */
        routeToView: function(viewId) {
            console.log('[ViewManager] Routing to view: ' + viewId);
            
            // Update menu state to trigger re-routing
            if (window.SmartDisplay.store) {
                window.SmartDisplay.store.setState({
                    menu: {
                        currentView: viewId
                    }
                });
            }
            
            // Trigger re-render
            this.render();
        },

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
         * 4. Alarm in alert or critical state → AlarmView (unless HA unconfigured with invalid alarm state)
         * 5. Menu.currentView = 'settings' and admin → SettingsView
         * 6. Otherwise → HomeView (or determined by menu.currentView)
         * 
         * Special case: When HA is not configured (haState.isConfigured === false) and alarm state
         * is invalid/waiting, route to Settings to allow user to configure HA (fixes first-boot lockout)
         * 
         * @param {object} state - Full application state
         * @returns {string} - View ID to display
         */
        getNextView: function(state) {
            // FAZ L1: Auth check - if not authenticated, show login
            var authState = state.authState || {};
            if (!authState.authenticated) {
                console.log('[ViewManager] Route: Login (not authenticated)');
                return 'login';
            }

            // Temporary override: once authenticated, bypass alarm-forced routing but still honor menu selection
            var bypassAlarmRouting = true;
            console.log('[ViewManager] Authenticated user → bypassing alarm routing');

            if (state.firstBoot) {
                console.log('[ViewManager] Route: FirstBoot');
                return 'first-boot';
            }

            var alarmState = state.alarmState || {};
            var haState = state.haState || {};

            console.log('[ViewManager] Routing debug - haState:', {
                isConfigured: haState.isConfigured,
                syncDone: haState.syncDone,
                fullState: haState
            });

            // Bypass alarm lock in these cases:
            // 1. HA not configured (first boot / not set up yet)
            // 2. HA configured but initial sync not done (ALWAYS bypass - can't trust alarm state without sync)
            // This prevents users from being locked to alarm screen during HA setup
            var haNotConfigured = !haState.isConfigured;
            var haSyncNotDone = haState.isConfigured && !haState.syncDone;
            
            // If HA sync is not done, bypass alarm lock regardless of alarm state
            if (haSyncNotDone) {
                console.log('[ViewManager] Bypassing alarm lock: HA sync not done yet');
            }
            
            var alarmStateInvalid = !alarmState.isHydrated || this._isAlarmStateInvalid(alarmState);
            var shouldBypassAlarmLock = haNotConfigured || haSyncNotDone || 
                                       (haNotConfigured && alarmStateInvalid);

            if (!bypassAlarmRouting) {
                if (!shouldBypassAlarmLock && (!alarmState.isHydrated || this._shouldLockToAlarm(alarmState))) {
                    console.log('[ViewManager] Route: Alarm (locked state)');
                    return 'alarm';
                }
            }

            if (state.menu && state.menu.currentView === 'menu') {
                console.log('[ViewManager] Route: Menu');
                return 'menu';
            }

            if (state.guestState && state.guestState.isGuestActive) {
                console.log('[ViewManager] Route: Guest');
                return 'guest';
            }

            // FAZ S0 & L1: Route protection - Settings Admin-only access
            if (state.menu && (state.menu.currentView === 'settings' || state.menu.currentView === 'homeassistant')) {
                var currentRole = authState.role || state.currentRole || 'guest';
                if (currentRole === 'admin') {
                    console.log('[ViewManager] Route: Settings');
                    return 'settings';
                } else {
                    console.log('[ViewManager] Route: Settings blocked for role: ' + currentRole + ', redirecting to Home');
                    return 'home';
                }
            }

            // Alarmo Settings view (admin/user accessible)
            // Map action IDs to view IDs
            var menuViewId = state.menu && state.menu.currentView;
            if (menuViewId === 'alarmo') {
                // Backend sends "alarmo" action ID, map to "alarmo-settings" view
                menuViewId = 'alarmo-settings';
            }

            if (menuViewId === 'alarmo-settings') {
                var currentRole = authState.role || state.currentRole || 'guest';
                if (currentRole === 'admin' || currentRole === 'user') {
                    console.log('[ViewManager] Route: Alarmo Settings');
                    return 'alarmo-settings';
                } else {
                    console.log('[ViewManager] Route: Alarmo Settings blocked for role: ' + currentRole + ', redirecting to Home');
                    return 'home';
                }
            }

            // Alarm Control view (admin/user accessible)
            if (state.menu && state.menu.currentView === 'alarm-control') {
                var currentRole = authState.role || state.currentRole || 'guest';
                if (currentRole === 'admin' || currentRole === 'user') {
                    console.log('[ViewManager] Route: Alarm Control');
                    return 'alarm-control';
                } else {
                    console.log('[ViewManager] Route: Alarm Control blocked for role: ' + currentRole + ', redirecting to Home');
                    return 'home';
                }
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
            
            // Prevent re-routing when already in Settings view with admin role
            // This keeps user in Settings during HA operations (test/sync)
            var currentRole = (state.authState || {}).role || state.currentRole || 'guest';
            var isInSettings = this.currentView && this.currentView.id === 'settings';
            var isAdmin = currentRole === 'admin';
            var settingsStillRequested = state.menu && state.menu.currentView === 'settings';
            
            if (isInSettings && isAdmin && settingsStillRequested) {
                // Stay in Settings, just update
                this._applyAlarmLock(state.alarmState || {}, state.haState || {});
                this._applyAuthMenuVisibility(state.authState || {});
                this._applyGuestModeIndicator(state.guestState || {}, state.authState || {});
                this.updateCurrentView(state);
                return;
            }
            
            var nextViewId = this.getNextView(state);

            this._applyAlarmLock(state.alarmState || {}, state.haState || {});

            // FAZ L1: Hide menu when not authenticated
            this._applyAuthMenuVisibility(state.authState || {});

            // FAZ L2: Update guest mode indicator visibility
            this._applyGuestModeIndicator(state.guestState || {}, state.authState || {});

            // Switch view if different
            if (!this.currentView || this.currentView.id !== nextViewId) {
                this.switchToView(nextViewId, state);
            } else {
                // Same view, update with new state
                this.updateCurrentView(state);
            }
        },

        _getAlarmoFields: function(alarmState) {
            var alarmo = (alarmState && alarmState.alarmo) || {};

            var state = alarmo.state || (alarmState && alarmState.state) || 'unknown';
            var triggered = (typeof alarmo.triggered === 'boolean') ? alarmo.triggered : Boolean(alarmState && alarmState.triggered);

            var delayRemaining = 0;
            if (alarmo.delay && typeof alarmo.delay.remaining === 'number') {
                delayRemaining = alarmo.delay.remaining;
            } else if (alarmState && alarmState.delay && typeof alarmState.delay.remaining === 'number') {
                delayRemaining = alarmState.delay.remaining;
            }

            return {
                state: state,
                triggered: triggered,
                delayRemaining: delayRemaining
            };
        },

        _logAlarmLockDecision: function(context, fields, locked) {
            console.log('[ViewManager] Alarm lock check (' + context + '):', {
                state: fields.state,
                triggered: fields.triggered,
                delayRemaining: fields.delayRemaining,
                locked: locked
            });
        },

        _shouldLockToAlarm: function(alarmState) {
            var alarmo = this._getAlarmoFields(alarmState);
            var state = (alarmo.state || 'unknown').toLowerCase();
            var triggered = alarmo.triggered;
            var delayRemaining = alarmo.delayRemaining;

            // Unlock when fully disarmed, not triggered, and no pending delay
            var unlocked = (state === 'disarmed') && !triggered && delayRemaining === 0;
            if (unlocked) {
                this._logAlarmLockDecision('shouldLock', alarmo, false);
                return false;
            }

            var locked = triggered || delayRemaining > 0 || state === 'arming' || state === 'pending' || state.indexOf('armed_') === 0;
            this._logAlarmLockDecision('shouldLock', alarmo, locked);
            return locked;
        },

        _isAlarmStateInvalid: function(alarmState) {
            var alarmo = this._getAlarmoFields(alarmState);
            var state = (alarmo.state || 'unknown').toLowerCase();

            // Invalid when we have no meaningful state yet
            if (!alarmState || !alarmState.isHydrated) {
                return true;
            }

            if (state === '' || state === 'unknown') {
                return true;
            }

            return false;
        },

        _applyAlarmLock: function(alarmState, haState) {
            // If user is authenticated, do not lock menu (temporary bypass)
            var authState = (window.SmartDisplay && window.SmartDisplay.store && window.SmartDisplay.store.getState().authState) || {};
            if (authState.authenticated) {
                this._logAlarmLockDecision('applyAlarmLock(authenticated bypass)', this._getAlarmoFields(alarmState), false);
                var overlayBypass = document.getElementById('menu-overlay');
                if (document && document.body) {
                    document.body.classList.remove('alarm-locked');
                }
                if (overlayBypass) {
                    overlayBypass.classList.remove('menu-locked');
                }
                return;
            }
            // First-boot bypass: don't lock menu if HA is not configured and alarm state is invalid
            var haNotConfigured = haState && !haState.isConfigured;
            var alarmStateInvalid = !alarmState || !alarmState.isHydrated || this._isAlarmStateInvalid(alarmState);
            var shouldBypassMenuLock = haNotConfigured && alarmStateInvalid;

            // Bypass lock while HA sync is pending
            var haSyncPending = haState && haState.isConfigured && haState.syncDone === false;

            var locked = !shouldBypassMenuLock && !haSyncPending && (!alarmState || !alarmState.isHydrated || this._shouldLockToAlarm(alarmState));
            this._logAlarmLockDecision('applyAlarmLock', this._getAlarmoFields(alarmState), locked);
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

        /**
         * Apply menu visibility based on auth state
         * FAZ L1: Hide menu when not authenticated
         */
        _applyAuthMenuVisibility: function(authState) {
            var authenticated = authState.authenticated || false;
            var overlay = document.getElementById('menu-overlay');

            if (overlay) {
                if (!authenticated) {
                    overlay.style.display = 'none';
                } else {
                    overlay.style.display = '';
                }
            }

            // Hide menu button when not authenticated
            if (document && document.body) {
                document.body.classList.toggle('auth-hidden', !authenticated);
            }
        },

        /**
         * Apply guest mode indicator visibility
         * FAZ L2: Show/hide guest mode indicator based on guest session state
         */
        _applyGuestModeIndicator: function(guestState, authState) {
            var isGuestActive = guestState.active && authState.role === 'guest';
            var indicator = document.getElementById('guest-mode-indicator');

            if (indicator) {
                indicator.style.display = isGuestActive ? 'flex' : 'none';
            }
        },

        isAlarmLocked: function() {
            if (!window.SmartDisplay || !window.SmartDisplay.store) {
                return false;
            }

            var state = window.SmartDisplay.store.getState();
            var authState = state.authState || {};
            var alarmState = state.alarmState;
            var haState = state.haState;

            // If user is authenticated, menu should not be blocked (temporary bypass)
            if (authState.authenticated) {
                this._logAlarmLockDecision('isAlarmLocked(authenticated bypass)', this._getAlarmoFields(alarmState), false);
                return false;
            }

            // Don't lock menu if HA sync is not done - user needs Settings access
            if (haState && haState.isConfigured && !haState.syncDone) {
                console.log('[ViewManager] Menu unlocked: HA sync pending');
                return false;
            }

            // First-boot bypass: don't lock if HA is not configured and alarm state is invalid
            var haNotConfigured = haState && !haState.isConfigured;
            var alarmStateInvalid = !alarmState || !alarmState.isHydrated || this._isAlarmStateInvalid(alarmState);
            var shouldBypassMenuLock = haNotConfigured && alarmStateInvalid;

            var locked = !shouldBypassMenuLock && (!alarmState || !alarmState.isHydrated || this._shouldLockToAlarm(alarmState));
            this._logAlarmLockDecision('isAlarmLocked', this._getAlarmoFields(alarmState), locked);
            return locked;
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
            
            // Pause HomeView inactivity timer while menu is open
            var homeView = this.views['home'];
            if (homeView && homeView.inactivityTimeout) {
                clearTimeout(homeView.inactivityTimeout);
                homeView.inactivityTimeout = null;
                console.log('[ViewManager] Paused HomeView inactivity timer');
            }
            
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
            
            // Resume HomeView inactivity timer when menu closes
            var homeView = this.views['home'];
            if (homeView && homeView._scheduleInactivityTimeout) {
                console.log('[ViewManager] Resuming HomeView inactivity timer');
                homeView._scheduleInactivityTimeout();
            }
        },

        /**
         * Reset menu auto-hide timer on user activity within menu
         * Called when user interacts with menu items (buttons, etc.)
         */
        resetMenuAutoHideTimer: function() {
            var homeView = this.views['home'];
            if (homeView && homeView._scheduleMenuAutoHide) {
                console.log('[ViewManager] Resetting menu auto-hide timer due to user activity');
                homeView._scheduleMenuAutoHide();
            }
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.viewManager = ViewManager;

    // Debug: Global countdown test access
    window.testCountdown = function(seconds) {
        if (ViewManager.views && ViewManager.views['alarm-control']) {
            ViewManager.views['alarm-control'].testCountdown(seconds);
        } else {
            console.log('🧪 AlarmControlView not active - navigate to alarm control first');
        }
    };

    // Auto-initialize when store is ready
    if (window.SmartDisplay.onReady) {
        window.SmartDisplay.onReady(function() {
            ViewManager.init();
        });
    }

    console.log('[SmartDisplay] View manager registered');

})();
