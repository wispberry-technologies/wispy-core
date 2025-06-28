// Main JavaScript file

// Wait for DOM to be ready
document.addEventListener('DOMContentLoaded', function() {
  console.log('Site initialized');
  
  // Example of theme toggling functionality
  const themeToggle = document.getElementById('theme-toggle');
  if (themeToggle) {
    themeToggle.addEventListener('click', function() {
      const html = document.querySelector('html');
      const currentTheme = html.getAttribute('data-theme');
      html.setAttribute('data-theme', currentTheme === 'light' ? 'dark' : 'light');
    });
  }
});
