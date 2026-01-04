/**
 * SmartDisplay First Boot Flow Controller
 * Manages first boot UI, API interactions, and state transitions
 */

(function() {
    'use strict';

    // ========================================================================
    // First Boot Controller
    // ========================================================================
    var FirstBootController = {
        currentStep: null,
        isLoading: false,
        error: null,

        // ====================================================================
        // Fetch Current Step
        // ====================================================================

        /**
         * Load first boot status from backend
         * @returns {Promise<object>} - Step data
         */
        fetchStatus: function() {
            var self = this;

            console.log('[FirstBoot] Fetching status...');

            return window.SmartDisplay.api.client.get('/setup/firstboot/status')
                .then(function(response) {
                    console.log('[FirstBoot] Status loaded:', response);
                    self.currentStep = response;
                    self.error = null;
                    return response;
                })
                .catch(function(err) {
                    console.error('[FirstBoot] Failed to fetch status:', err);
                    self.error = err.message || 'Failed to load first boot status';
                    throw err;
                });
        },

        // ====================================================================
        // Step Navigation
        // ====================================================================

        /**
         * Move to next step
         * @returns {Promise<object>} - Next step data
         */
        nextStep: function() {
            var self = this;

            if (this.isLoading) {
                console.warn('[FirstBoot] Already loading, skipping next');
                return Promise.reject({ message: 'Already processing' });
            }

            this.isLoading = true;
            console.log('[FirstBoot] Moving to next step...');

            return window.SmartDisplay.api.client.post('/setup/firstboot/next', {})
                .then(function(response) {
                    console.log('[FirstBoot] Next step loaded:', response);
                    self.currentStep = response;
                    self.error = null;
                    self.isLoading = false;
                    
                    // Update store with new state
                    window.SmartDisplay.store.setState({ firstBoot: true });
                    
                    return response;
                })
                .catch(function(err) {
                    console.error('[FirstBoot] Failed to move to next step:', err);
                    self.error = err.message || 'Failed to move to next step';
                    self.isLoading = false;
                    throw err;
                });
        },

        /**
         * Move to previous step
         * @returns {Promise<object>} - Previous step data
         */
        previousStep: function() {
            var self = this;

            if (this.isLoading) {
                console.warn('[FirstBoot] Already loading, skipping back');
                return Promise.reject({ message: 'Already processing' });
            }

            this.isLoading = true;
            console.log('[FirstBoot] Moving to previous step...');

            return window.SmartDisplay.api.client.post('/setup/firstboot/back', {})
                .then(function(response) {
                    console.log('[FirstBoot] Previous step loaded:', response);
                    self.currentStep = response;
                    self.error = null;
                    self.isLoading = false;
                    return response;
                })
                .catch(function(err) {
                    console.error('[FirstBoot] Failed to move to previous step:', err);
                    self.error = err.message || 'Failed to move to previous step';
                    self.isLoading = false;
                    throw err;
                });
        },

        /**
         * Complete first boot
         * @returns {Promise<object>} - Completion response
         */
        complete: function() {
            var self = this;

            if (this.isLoading) {
                console.warn('[FirstBoot] Already loading, skipping complete');
                return Promise.reject({ message: 'Already processing' });
            }

            this.isLoading = true;
            console.log('[FirstBoot] Completing first boot...');

            return window.SmartDisplay.api.client.post('/setup/firstboot/complete', {})
                .then(function(response) {
                    console.log('[FirstBoot] First boot completed:', response);
                    self.currentStep = null;
                    self.error = null;
                    self.isLoading = false;
                    
                    // Update store - first boot complete
                    window.SmartDisplay.store.setState({ firstBoot: false });
                    
                    return response;
                })
                .catch(function(err) {
                    console.error('[FirstBoot] Failed to complete first boot:', err);
                    self.error = err.message || 'Failed to complete first boot';
                    self.isLoading = false;
                    throw err;
                });
        },

        // ====================================================================
        // Initialization
        // ====================================================================

        /**
         * Initialize controller and load initial step
         */
        init: function() {
            var self = this;

            console.log('[FirstBoot] Initializing controller');

            // Load initial status
            return this.fetchStatus()
                .catch(function(err) {
                    console.error('[FirstBoot] Failed to initialize:', err);
                    // Continue anyway - will show error in view
                });
        }
    };

    // Add POST method to API client if not exists
    if (!window.SmartDisplay.api.client.post) {
        window.SmartDisplay.api.client.post = function(endpoint, data, options) {
            options = options || {};
            var timeout = options.timeout || window.SmartDisplay.api.timeout;
            var requestId = 'req-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);

            var url = endpoint.startsWith('http') 
                ? endpoint 
                : window.SmartDisplay.api.baseUrl + endpoint;

            console.log('[API] POST ' + endpoint + ' (ID: ' + requestId + ')');

            return new Promise(function(resolve, reject) {
                var xhr = new XMLHttpRequest();
                var timeoutHandle;

                timeoutHandle = setTimeout(function() {
                    xhr.abort();
                    console.error('[API] Timeout ' + endpoint + ' (ID: ' + requestId + ')');
                    reject({
                        type: 'TIMEOUT',
                        requestId: requestId,
                        endpoint: endpoint,
                        message: 'Request timeout after ' + timeout + 'ms'
                    });
                }, timeout);

                xhr.open('POST', url, true);
                xhr.setRequestHeader('X-Request-ID', requestId);
                xhr.setRequestHeader('Content-Type', 'application/json');
                xhr.setRequestHeader('Accept', 'application/json');

                xhr.onreadystatechange = function() {
                    if (xhr.readyState === 4) {
                        clearTimeout(timeoutHandle);

                        var response;
                        try {
                            response = xhr.responseText ? JSON.parse(xhr.responseText) : {};
                        } catch (e) {
                            console.error('[API] JSON parse error ' + endpoint + ' (ID: ' + requestId + ')');
                            return reject({
                                type: 'PARSE_ERROR',
                                requestId: requestId,
                                endpoint: endpoint,
                                message: 'Failed to parse response'
                            });
                        }

                        if (xhr.status >= 200 && xhr.status < 300) {
                            if (response && response.error && typeof response.error === 'string') {
                                console.error('[API] Error envelope ' + endpoint + ' (ID: ' + requestId + '): ' + response.error);
                                return reject({
                                    type: 'API_ERROR',
                                    requestId: requestId,
                                    endpoint: endpoint,
                                    message: response.error,
                                    code: response.code || null
                                });
                            }

                            console.log('[API] Success ' + endpoint + ' (ID: ' + requestId + ')');
                            resolve(response);
                        } else {
                            console.error('[API] HTTP ' + xhr.status + ' ' + endpoint + ' (ID: ' + requestId + ')');
                            reject({
                                type: 'HTTP_ERROR',
                                requestId: requestId,
                                endpoint: endpoint,
                                statusCode: xhr.status,
                                statusText: xhr.statusText,
                                message: 'HTTP ' + xhr.status + ': ' + xhr.statusText,
                                response: response
                            });
                        }
                    }
                };

                xhr.onerror = function() {
                    clearTimeout(timeoutHandle);
                    console.error('[API] Network error ' + endpoint + ' (ID: ' + requestId + ')');
                    reject({
                        type: 'NETWORK_ERROR',
                        requestId: requestId,
                        endpoint: endpoint,
                        message: 'Network error or CORS blocked'
                    });
                };

                var jsonData = JSON.stringify(data || {});
                xhr.send(jsonData);
            });
        };
    }

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.firstBootController = FirstBootController;

    console.log('[SmartDisplay] First Boot controller registered');

})();
