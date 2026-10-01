(() => {
    const authForms = document.getElementById("auth-forms");
    const authenticated = document.getElementById("already-authenticated");
    if (API.isAuthenticated) {
        authForms?.classList.add("hidden");
        authenticated?.classList.remove("hidden");
    }

    const loginForm = document.getElementById("login-form");
    const signupForm = document.getElementById("signup-form");

    loginForm?.addEventListener("submit", async (event) => {
        event.preventDefault();
        const error = document.getElementById("login-error");
        error.textContent = "";
        try {
            const login = document.getElementById("login").value.trim();
            const password = document.getElementById("password").value;
            await API.login(login, password);
            window.location.href = "/play";
        } catch (e) {
            error.textContent = e.message || "Не удалось войти";
        }
    });

    signupForm?.addEventListener("submit", async (event) => {
        event.preventDefault();
        const error = document.getElementById("signup-error");
        const success = document.getElementById("signup-success");
        error.textContent = "";
        success.textContent = "";
        try {
            const login = document.getElementById("signup-login").value.trim();
            const password = document.getElementById("signup-password").value;
            await API.signUp(login, password);
            success.textContent = "Пользователь создан. Выполняю вход...";
            await API.login(login, password);
            window.location.href = "/play";
        } catch (e) {
            error.textContent = e.message || "Не удалось зарегистрироваться";
        }
    });
})();
