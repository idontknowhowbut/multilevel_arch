(() => {
    const loginLink = document.getElementById("login-link");
    const logoutButton = document.getElementById("logout-button");
    const title = document.getElementById("session-title");
    const user = document.getElementById("session-user");
    const dot = document.querySelector(".session-status .status-dot");

    if (!loginLink || !logoutButton || !window.API) return;

    if (API.isAuthenticated) {
        loginLink.classList.add("hidden");
        logoutButton.classList.remove("hidden");
        title.textContent = "ONLINE";
        user.textContent = API.loginName || "Авторизован";
        dot?.classList.add("status-dot--online");
    } else {
        loginLink.classList.remove("hidden");
        logoutButton.classList.add("hidden");
    }

    logoutButton.addEventListener("click", () => {
        API.clearSession();
        window.location.href = "/login";
    });
})();
