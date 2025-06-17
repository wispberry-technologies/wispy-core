// Main application JavaScript
console.log('Main app.js loaded');

// Demo functionality
document.addEventListener('DOMContentLoaded', function() {
    console.log('DOM Content Loaded - app.js');
    
    // Add hover effects to cards
    const cards = document.querySelectorAll('.card');
    cards.forEach(card => {
        card.classList.add('card-hover');
    });
    
    // Demo button interactions
    const demoButtons = document.querySelectorAll('.demo-button');
    demoButtons.forEach(button => {
        button.addEventListener('click', function() {
            console.log('Demo button clicked');
            this.textContent = 'Clicked!';
            setTimeout(() => {
                this.textContent = 'Demo Button';
            }, 2000);
        });
    });
});
