javascript:(function() {
    const token = document.cookie.split('; ').find(row => row.startsWith('WorkosCursorSessionToken='));
    if (token) {
        const tokenValue = token.split('=')[1];
        prompt('Your Cursor session token (WorkosCursorSessionToken):', tokenValue);
    } else {
        alert('WorkosCursorSessionToken not found. Please make sure you are logged in to Cursor.');
    }
})();

