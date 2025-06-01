// Function to handle API requests with token refresh
async function fetchWithAuth(url, options = {}) {
    // Try the request with current token
    let response = await fetch(url, options);
    
    // If unauthorized, try to refresh the token
    if (response.status === 401 || response.status === 403) {
        const refreshResponse = await fetch('/refresh', {
            method: 'POST',
            credentials: 'include' // Include cookies
        });
        
        // If refresh successful, retry the original request
        if (refreshResponse.ok) {
            return fetch(url, options);
        } else {
            // If refresh failed, redirect to login
            window.location.href = '/signin';
            return null;
        }
    }
    
    return response;
}

// Add event listener to handle page load
document.addEventListener('DOMContentLoaded', function() {
    // Add logout functionality to logout buttons
    const logoutButtons = document.querySelectorAll('.logout-btn');
    if (logoutButtons) {
        logoutButtons.forEach(button => {
            button.addEventListener('click', function(e) {
                e.preventDefault();
                window.location.href = '/logout';
            });
        });
    }
});