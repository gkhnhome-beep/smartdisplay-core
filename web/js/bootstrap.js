/**
 * SmartDisplay Frontend Bootstrap
 * Initializes global state and application foundation
 * Kiosk-safe configuration - no external links or navigation outside app
 */

(function() {
    'use strict';

    // ========================================================================
    // Global Application State
    // ========================================================================
    window.SmartDisplay = {
        // API Configuration
        api: {
            baseUrl: 'http://localhost:8090/api',
            timeout: 30000
        },

        // Application State
        state: {
            isInitialized: false,
            isOnline: true,
            currentUser: null,
            appVersion: '1.0.0',
            currentRole: 'admin'  // FAZ S0: Role-based access control (admin|user|guest)
        },

        // Configuration
        config: {
            enableTelemetry: true,
            enableLogging: true,
            maxRetries: 3
        },

        // A6.C: UI Strings Dictionary - Centralized for maintainability and future i18n
        strings: {
            alarm: {
                triggered: 'ALARM LOCKDOWN IN EFFECT',
                arming: 'Arming in progress',
                pending: 'Entry/exit pending',
                armed: 'System armed',
                disarmed: 'System disarmed',
                unknown: 'System state',
                
                explanation: {
                    triggered: 'Alarm triggered.',
                    arming: 'Exit delay active.',
                    pending: 'Entry delay active.',
                    armed: 'Alarm is armed. Disarm required to continue.',
                    disarmed: ''
                },
                
                action: {
                    sending: 'Sending request...',
                    waiting: 'Waiting for alarm state update\u2026',
                    blocked: 'Alarm is triggered. Action blocked.',
                    unreachable: 'Alarm system unreachable.',
                    invalid: 'Invalid request.',
                    failed: 'Action request failed'
                }
            },
            
            connection: {
                retrying: 'Connection issue. Retrying\u2026',
                loading: 'Waiting for Alarmo state...',
                connecting: 'Connecting...'
            },
            
            error: {
                controllerNotInit: 'Controller not initialized'
            }
        },

        // A6.D: Diagnostic Mode - Runtime flag for extra logging (default: OFF)
        diagnostic: {
            enabled: false,
            log: function(component, message, data) {
                if (this.enabled) {
                    var logMsg = '[' + component + '] ' + message;
                    if (data !== undefined) {
                        console.log(logMsg, data);
                    } else {
                        console.log(logMsg);
                    }
                }
            }
        },

        // A6.B: Kiosk Longevity Features
        kiosk: {
            idleTime: 0,
            idleThreshold: 300000, // 5 minutes
            burnInShiftInterval: null,
            burnInShiftActive: false,
            
            startBurnInPrevention: function() {
                if (this.burnInShiftActive) return;
                
                var self = this;
                this.burnInShiftInterval = setInterval(function() {
                    // Only shift during idle (and respect reduced-motion)
                    if (self.idleTime < self.idleThreshold) return;
                    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
                    
                    var shiftX = Math.sin(Date.now() / 30000) * 2;
                    var shiftY = Math.cos(Date.now() / 30000) * 2;
                    
                    var appEl = document.getElementById('app');
                    if (appEl) {
                        appEl.style.transform = 'translate(' + shiftX + 'px, ' + shiftY + 'px)';
                    }
                }, 10000);
                
                this.burnInShiftActive = true;
            },
            
            stopBurnInPrevention: function() {
                if (this.burnInShiftInterval) {
                    clearInterval(this.burnInShiftInterval);
                    this.burnInShiftInterval = null;
                }
                
                var appEl = document.getElementById('app');
                if (appEl) {
                    appEl.style.transform = '';
                }
                
                this.burnInShiftActive = false;
            },
            
            updateIdleState: function() {
                var body = document.body;
                if (this.idleTime >= this.idleThreshold) {
                    body.classList.add('long-idle');
                } else {
                    body.classList.remove('long-idle');
                }
            },
            
            resetIdle: function() {
                this.idleTime = 0;
                this.updateIdleState();
            },
            
            incrementIdle: function(ms) {
                this.idleTime += ms;
                this.updateIdleState();
            }
        },

        // Lifecycle Hooks
        hooks: {
            onInit: [],
            onReady: [],
            onDestroy: []
        }
    };

    // ========================================================================
    // Kiosk Safety: Disable Context Menu
    // ========================================================================
    document.addEventListener('contextmenu', function(e) {
        e.preventDefault();
        return false;
    }, false);

    // ========================================================================
    // Kiosk Safety: Prevent Right-Click Touch Menu
    // ========================================================================
    document.addEventListener('touchstart', function(e) {
        if (e.touches.length > 1) {
            e.preventDefault();
        }
    }, false);

    // ========================================================================
    // Kiosk Safety: Trap Navigation - Prevent Leaving App
    // ========================================================================
    document.addEventListener('click', function(e) {
        var target = e.target.closest('a');
        
        if (target) {
            var href = target.getAttribute('href');
            
            // Allow internal navigation (hash-based or relative paths within /app)
            if (href && !href.startsWith('#') && 
                !href.startsWith('/app') && 
                !href.startsWith('javascript:')) {
                
                // Block external navigation
                if (href.startsWith('http://') || href.startsWith('https://')) {
                    e.preventDefault();
                    return false;
                }
            }
        }
    }, false);

    // ========================================================================
    // Kiosk Safety: Prevent Zoom
    // ========================================================================
    document.addEventListener('wheel', function(e) {
        if (e.ctrlKey) {
            e.preventDefault();
        }
    }, { passive: false });

    // Prevent pinch zoom
    document.addEventListener('touchmove', function(e) {
        if (e.touches.length > 1) {
            e.preventDefault();
        }
    }, { passive: false });

    // ========================================================================
    // Application Initialization
    // ========================================================================
    function init() {
        console.log('[SmartDisplay] Bootstrapping application...');

        // API runs on fixed backend port (localhost:8090)
        window.SmartDisplay.api.baseUrl = 'http://localhost:8090/api';

        // Log resolved API base URL
        console.log('[SmartDisplay] API Base URL: ' + window.SmartDisplay.api.baseUrl);

        // Mark as initialized
        window.SmartDisplay.state.isInitialized = true;

        // Execute init hooks
        if (window.SmartDisplay.hooks.onInit.length > 0) {
            window.SmartDisplay.hooks.onInit.forEach(function(hook) {
                try {
                    hook();
                } catch (e) {
                    console.error('[SmartDisplay] Init hook error:', e);
                }
            });
        }

        // Dispatch ready event
        dispatchReady();
    }

    // ========================================================================
    // Ready State
    // ========================================================================
    function dispatchReady() {
        console.log('[SmartDisplay] Application ready');
        console.log('[SmartDisplay] API Base URL:', window.SmartDisplay.api.baseUrl);

        // Initialize Cinematic Alarm Scene System
        if (window.SmartDisplay.AlarmScene) {
            window.SmartDisplay.AlarmScene.init();
        }

        // FAZ L4: Initialize Advisor
        if (window.SmartDisplay.advisor) {
            window.SmartDisplay.advisor.init();
        }

        // FAZ L6: Initialize Trace
        if (window.SmartDisplay.trace) {
            window.SmartDisplay.trace.init();
        }

        // FAZ L5: Play intro if first boot
        if (window.SmartDisplay.intro && window.SmartDisplay.intro.shouldShow()) {
            console.log('[Bootstrap] First boot detected, playing intro...');
            window.SmartDisplay.intro.play().catch(function(e) {
                console.error('[Bootstrap] Intro error (continuing):', e);
            });
        }

        // Execute ready hooks
        if (window.SmartDisplay.hooks.onReady.length > 0) {
            window.SmartDisplay.hooks.onReady.forEach(function(hook) {
                try {
                    hook();
                } catch (e) {
                    console.error('[SmartDisplay] Ready hook error:', e);
                }
            });
        }

        // Dispatch custom ready event
        var readyEvent = new CustomEvent('smartdisplay-ready', {
            detail: { state: window.SmartDisplay.state }
        });
        document.dispatchEvent(readyEvent);
    }

    // ========================================================================
    // Register Hook Methods
    // ========================================================================
    window.SmartDisplay.onInit = function(callback) {
        if (typeof callback === 'function') {
            window.SmartDisplay.hooks.onInit.push(callback);
        }
    };

    window.SmartDisplay.onReady = function(callback) {
        if (typeof callback === 'function') {
            window.SmartDisplay.hooks.onReady.push(callback);
        }
    };

    window.SmartDisplay.onDestroy = function(callback) {
        if (typeof callback === 'function') {
            window.SmartDisplay.hooks.onDestroy.push(callback);
        }
    };

    // ========================================================================
    // Cleanup on Page Unload
    // ========================================================================
    window.addEventListener('beforeunload', function() {
        // Cleanup AlarmScene
        if (window.SmartDisplay.AlarmScene) {
            window.SmartDisplay.AlarmScene.destroy();
        }

        window.SmartDisplay.hooks.onDestroy.forEach(function(hook) {
            try {
                hook();
            } catch (e) {
                console.error('[SmartDisplay] Destroy hook error:', e);
            }
        });
    });

    // ========================================================================
    // Start Bootstrap when DOM is Ready
    // ========================================================================
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

    // ========================================================================
    // A6.B: Initialize Kiosk Longevity Features
    // ========================================================================
    window.SmartDisplay.onReady(function() {
        window.SmartDisplay.kiosk.startBurnInPrevention();
        
        // Track idle time
        setInterval(function() {
            window.SmartDisplay.kiosk.incrementIdle(1000);
        }, 1000);
        
        // Reset idle on any interaction
        ['mousedown', 'mousemove', 'keydown', 'touchstart', 'scroll'].forEach(function(eventType) {
            document.addEventListener(eventType, function() {
                window.SmartDisplay.kiosk.resetIdle();
            }, { passive: true });
        });
    });

    // ========================================================================
    // A6.A: System Status Overlay (Ctrl+Shift+D)\n    // ========================================================================
    window.SmartDisplay.onReady(function() {
        var overlay = document.getElementById('system-status-overlay');
        var closeBtn = document.getElementById('status-close-btn');
        
        function updateSystemStatus() {
            // Fetch health data
            fetch(window.SmartDisplay.api.baseUrl.replace('/api', '') + '/health')
                .then(function(r) { return r.json(); })
                .then(function(health) {
                    document.getElementById('status-version').textContent = health.version || '1.0.0-rc1';
                    document.getElementById('status-backend').textContent = health.status === 'ok' ? 'Connected' : 'Degraded';
                    document.getElementById('status-alarmo').textContent = health.ha_connected ? 'Connected' : 'Disconnected';
                    
                    var controller = window.SmartDisplay.alarmController;
                    if (controller && controller.lastUpdateTime) {
                        var ago = Math.floor((Date.now() - controller.lastUpdateTime) / 1000);
                        document.getElementById('status-last-poll').textContent = ago + 's ago';
                    } else {
                        document.getElementById('status-last-poll').textContent = 'Never';
                    }
                })
                .catch(function() {
                    document.getElementById('status-backend').textContent = 'Unreachable';
                });
        }
        
        function toggleSystemStatus() {
            if (overlay.style.display === 'none') {
                updateSystemStatus();
                overlay.style.display = 'flex';
            } else {
                overlay.style.display = 'none';
            }
        }
        
        // Ctrl+Shift+D to toggle
        document.addEventListener('keydown', function(e) {
            if (e.ctrlKey && e.shiftKey && e.key === 'D') {
                e.preventDefault();
                toggleSystemStatus();
            }
        });
        
        if (closeBtn) {
            closeBtn.addEventListener('click', function() {
                overlay.style.display = 'none';
            });
        }
        
        // Close on overlay click (not content)
        if (overlay) {
            overlay.addEventListener('click', function(e) {
                if (e.target === overlay) {
                    overlay.style.display = 'none';
                }
            });
        }
    });

    // ========================================================================
    // TURKISH COUNTDOWN TEST - Debug function for user issue
    // ========================================================================
    window.testTurkishCountdown = function(seconds) {
        seconds = seconds || 30; // Default 30 seconds as user requested
        console.log('[Bootstrap] 🇹🇷 Starting Turkish countdown test with', seconds, 'seconds');
        
        // Get AlarmControlView directly
        var alarmView = window.SmartDisplay.viewManager ? window.SmartDisplay.viewManager.views.filter(function(v) { 
            return v.id === 'alarm-control'; 
        })[0] : null;
        
        if (alarmView && alarmView._showCountdownOverlay) {
            console.log('[Bootstrap] Found AlarmControlView, showing countdown...');
            alarmView._showCountdownOverlay(seconds);
        } else {
            console.error('[Bootstrap] AlarmControlView not found or no countdown method');
            console.log('[Bootstrap] Available views:', window.SmartDisplay.viewManager ? window.SmartDisplay.viewManager.views : 'No viewManager');
        }
    };
    
    // Quick test function - just shows countdown immediately
    window.quickCountdownTest = function() {
        console.log('🇹🇷 Quick countdown test starting...');
        
        // Direct DOM manipulation test first
        var overlay = document.getElementById('alarm-overlay');
        var countdownOverlay = document.getElementById('alarm-countdown-overlay');
        var countdownValue = document.getElementById('alarm-countdown-value');
        
        console.log('DOM elements found:', {
            overlay: !!overlay,
            countdownOverlay: !!countdownOverlay, 
            countdownValue: !!countdownValue
        });
        
        if (overlay && countdownOverlay && countdownValue) {
            console.log('✅ All DOM elements found, starting direct countdown...');
            
            // Show overlay directly
            overlay.classList.add('active');
            overlay.style.setProperty('--pulse-speed', '2.5s');
            countdownOverlay.classList.add('active');
            
            // Start countdown from 30
            var remaining = 30;
            countdownValue.textContent = String(remaining).padStart(2, '0');
            
            var timer = setInterval(function() {
                remaining--;
                console.log('⏰ Countdown:', remaining);
                
                if (remaining <= 0) {
                    clearInterval(timer);
                    overlay.classList.remove('active');
                    countdownOverlay.classList.remove('active');
                    console.log('⏰ Countdown finished!');
                    return;
                }
                
                // Fast pulse for last 10 seconds
                if (remaining <= 10) {
                    overlay.style.setProperty('--pulse-speed', '1.2s');
                    countdownOverlay.classList.add('warning');
                }
                
                countdownValue.textContent = String(remaining).padStart(2, '0');
            }, 1000);
            
        } else {
            console.error('❌ Missing DOM elements - need to be on alarm page first');
            console.log('Current page elements:', document.querySelectorAll('[id*="alarm"]').length);
        }
    };
    
    // Force show alarm view and start countdown
    window.forceAlarmCountdown = function() {
        console.log('🚀 Force starting alarm countdown...');
        
        // Navigate to alarm view first
        if (window.SmartDisplay && window.SmartDisplay.viewManager) {
            window.SmartDisplay.viewManager.showView('alarm-control');
            
            // Wait a bit for DOM to be ready
            setTimeout(function() {
                console.log('📍 Now on alarm page, starting countdown...');
                quickCountdownTest();
            }, 500);
        } else {
            console.error('❌ SmartDisplay not initialized');
        }
    };

})();
