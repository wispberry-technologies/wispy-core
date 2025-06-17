// Initialization JavaScript for immediate execution
(function() {
    'use strict';
    
    // Initialize critical functionality
    console.log('Wispy Core: Critical initialization started');
    
    // Set up viewport meta tag for responsive design
    const viewport = document.querySelector('meta[name="viewport"]');
    if (!viewport) {
        const meta = document.createElement('meta');
        meta.name = 'viewport';
        meta.content = 'width=device-width, initial-scale=1.0';
        document.head.appendChild(meta);
    }
    
    // Initialize theme detection
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');
    const currentTheme = localStorage.getItem('theme') || (prefersDark.matches ? 'dark' : 'light');
    document.documentElement.setAttribute('data-theme', currentTheme);
    
    // Initialize loading state
    document.documentElement.classList.add('wispy-loading');
    
    // Remove loading state when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', function() {
            document.documentElement.classList.remove('wispy-loading');
            console.log('Wispy Core: DOM ready, initialization complete');
        });
    } else {
        document.documentElement.classList.remove('wispy-loading');
        console.log('Wispy Core: Initialization complete');
    }
    
    // Initialize error handling
    window.WispyCore = window.WispyCore || {};
    window.WispyCore.initialized = true;
    
})();
