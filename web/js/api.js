/**
 * SmartDisplay API Client
 * Handles all backend communication with request logging and error handling
 */

(function() {
    'use strict';

    // ========================================================================
    // Request ID Generator
    // ========================================================================
    var requestIdCounter = 0;

    function generateRequestId() {
        requestIdCounter++;
        return 'req-' + Date.now() + '-' + requestIdCounter;
    }

    // ========================================================================
    // API Client
    // ========================================================================
    var ApiClient = {
        /**
         * Perform GET request to API
         * @param {string} endpoint - API endpoint (e.g., '/health', '/api/alarm/status')
         * @param {object} options - Optional configuration
         * @param {number} options.timeout - Request timeout in ms (default: 30000)
         * @param {function} options.onTimeout - Callback on timeout
         * @param {object} options.headers - Additional headers to send (e.g., {'X-User-Role': 'admin'})
         * @returns {Promise<object>} - Resolves with response data, rejects with error
         */
        get: function(endpoint, options) {
            options = options || {};
            var timeout = options.timeout || window.SmartDisplay.api.timeout;
            var customHeaders = options.headers || {};
            var requestId = generateRequestId();

            // Ensure endpoint is absolute
            var url = endpoint.startsWith('http') 
                ? endpoint 
                : window.SmartDisplay.api.baseUrl + endpoint;

            console.log('[API] GET ' + endpoint + ' (ID: ' + requestId + ')');

            return new Promise(function(resolve, reject) {
                var xhr = new XMLHttpRequest();
                var timeoutHandle;

                // Setup timeout
                timeoutHandle = setTimeout(function() {
                    xhr.abort();
                    var timeoutError = {
                        type: 'TIMEOUT',
                        requestId: requestId,
                        endpoint: endpoint,
                        message: 'Request timeout after ' + timeout + 'ms'
                    };
                    console.error('[API] Timeout ' + endpoint + ' (ID: ' + requestId + ')');
                    if (typeof options.onTimeout === 'function') {
                        options.onTimeout(timeoutError);
                    }
                    reject(timeoutError);
                }, timeout);

                // Setup request
                xhr.open('GET', url, true);
                xhr.setRequestHeader('X-Request-ID', requestId);
                xhr.setRequestHeader('Accept', 'application/json');

                // Add custom headers
                for (var headerName in customHeaders) {
                    if (customHeaders.hasOwnProperty(headerName)) {
                        xhr.setRequestHeader(headerName, customHeaders[headerName]);
                    }
                }

                // Handle response
                xhr.onreadystatechange = function() {
                    if (xhr.readyState === 4) {
                        clearTimeout(timeoutHandle);

                        // Parse response
                        var response;
                        try {
                            response = xhr.responseText ? JSON.parse(xhr.responseText) : {};
                        } catch (e) {
                            console.error('[API] JSON parse error ' + endpoint + ' (ID: ' + requestId + ')');
                            return reject({
                                type: 'PARSE_ERROR',
                                requestId: requestId,
                                endpoint: endpoint,
                                message: 'Failed to parse response',
                                originalError: e
                            });
                        }

                        // Check HTTP status
                        if (xhr.status >= 200 && xhr.status < 300) {
                            // Success - but check envelope
                            if (ApiClient._isErrorEnvelope(response)) {
                                console.error('[API] Error envelope ' + endpoint + ' (ID: ' + requestId + '): ' + response.error);
                                return reject(ApiClient._normalizeError(response, requestId, endpoint));
                            }

                            console.log('[API] Success ' + endpoint + ' (ID: ' + requestId + ')');
                            resolve(response);
                        } else {
                            // HTTP error
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

                // Handle network errors
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

                xhr.send();
            });
        },

        /**
         * Check if response is an error envelope from backend
         * @private
         */
        _isErrorEnvelope: function(response) {
            return response && 
                   response.error && 
                   typeof response.error === 'string' && 
                   response.error.length > 0;
        },

        /**
         * Normalize error response into consistent error object
         * @private
         */
        _normalizeError: function(response, requestId, endpoint) {
            var error = {
                type: 'API_ERROR',
                requestId: requestId,
                endpoint: endpoint,
                message: response.error || 'Unknown error',
                code: response.code || null,
                details: response.details || null
            };

            // Include data if present (useful for validation errors)
            if (response.data) {
                error.data = response.data;
            }

            return error;
        }
    };

    // ========================================================================
    // Register with SmartDisplay Global
    // ========================================================================
    window.SmartDisplay.api.client = ApiClient;

    console.log('[SmartDisplay] API client registered');

})();
