// Analytics simulation
console.log('Analytics script loaded');

// Simulate analytics tracking
window.trackEvent = function(event, data) {
    console.log('Analytics Event:', event, data);
};

// Track page view
window.trackEvent('page_view', {
    page: window.location.pathname,
    timestamp: new Date().toISOString()
});
