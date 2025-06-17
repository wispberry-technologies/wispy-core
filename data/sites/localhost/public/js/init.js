// Initialization script - should be inlined
console.log('Initialization script running');

// Set up global configuration
window.WISPY_CONFIG = {
    version: '1.0.0',
    debug: true,
    features: {
        importDemo: true
    }
};

// Early initialization
(function() {
    console.log('Early initialization complete');
})();
