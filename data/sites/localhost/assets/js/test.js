// Test JavaScript file for assets
console.log('Asset test JS loaded');

function assetTestFunction() {
    console.log('Asset test function called');
    return 'Asset test success';
}

window.ASSET_TEST = {
    loaded: true,
    timestamp: new Date().toISOString()
};
