(() => {
    const ACCESS_KEY = "ttt.accessToken";
    const REFRESH_KEY = "ttt.refreshToken";
    const LOGIN_KEY = "ttt.login";

    async function errorText(response) {
        const text = (await response.text()).trim();
        return text || `HTTP ${response.status}`;
    }

    const API = {
        get accessToken() { return localStorage.getItem(ACCESS_KEY); },
        get refreshToken() { return localStorage.getItem(REFRESH_KEY); },
        get loginName() { return localStorage.getItem(LOGIN_KEY); },
        get isAuthenticated() { return Boolean(this.accessToken); },

        saveSession(tokens, login) {
            localStorage.setItem(ACCESS_KEY, tokens.accessToken);
            localStorage.setItem(REFRESH_KEY, tokens.refreshToken);
            if (login) localStorage.setItem(LOGIN_KEY, login);
        },

        clearSession() {
            localStorage.removeItem(ACCESS_KEY);
            localStorage.removeItem(REFRESH_KEY);
            localStorage.removeItem(LOGIN_KEY);
            localStorage.removeItem("ttt.currentGameId");
            localStorage.removeItem("ttt.currentGameOwnerId");
            localStorage.removeItem("ttt.currentGameWaiting");
        },

        requireAuth() {
            if (this.isAuthenticated) return true;
            window.location.href = "/login";
            return false;
        },

        async login(login, password) {
            const response = await fetch("/auth/login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ login, password }),
            });
            if (!response.ok) throw new Error(await errorText(response));
            const tokens = await response.json();
            this.saveSession(tokens, login);
            return tokens;
        },

        async signUp(login, password) {
            const response = await fetch("/signUp", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ login, password }),
            });
            if (!response.ok) throw new Error(await errorText(response));
        },

        async refresh() {
            if (!this.refreshToken) throw new Error("Refresh token отсутствует");
            const response = await fetch("/auth/refresh", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ refreshToken: this.refreshToken }),
            });
            if (!response.ok) throw new Error(await errorText(response));
            const tokens = await response.json();
            this.saveSession(tokens, this.loginName);
            return tokens.accessToken;
        },

        async request(url, options = {}, retry = true) {
            if (!this.accessToken) throw new Error("Требуется авторизация");
            const headers = { ...(options.headers || {}) };
            headers.Authorization = `Bearer ${this.accessToken}`;
            if (options.body && !(options.body instanceof FormData) && !headers["Content-Type"]) {
                headers["Content-Type"] = "application/json";
            }

            let response = await fetch(url, { ...options, headers });
            if (response.status === 401 && retry && this.refreshToken) {
                try {
                    const accessToken = await this.refresh();
                    headers.Authorization = `Bearer ${accessToken}`;
                    response = await fetch(url, { ...options, headers });
                } catch (_) {
                    this.clearSession();
                    window.location.href = "/login";
                }
            }
            return response;
        },

        errorText,
    };

    window.API = API;
})();
