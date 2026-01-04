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
            appVersion: '1.0.0'
        },

        // Configuration
        config: {
            enableTelemetry: true,
            enableLogging: true,
            maxRetries: 3
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

})();
