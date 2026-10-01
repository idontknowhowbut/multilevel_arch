(() => {
    const ACCESS = "ttt.accessToken";
    const REFRESH = "ttt.refreshToken";
    const LOGIN = "ttt.login";
    const GAME = "ttt.currentGameId";

    const $ = id => document.getElementById(id);
    const esc = value => String(value ?? "").replace(/[&<>"']/g, c => ({
        "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;"
    }[c]));

    async function message(response) {
        const text = (await response.text()).trim();
        return text || `HTTP ${response.status}`;
    }

    const api = {
        get token() { return localStorage.getItem(ACCESS); },
        get refreshToken() { return localStorage.getItem(REFRESH); },
        get login() { return localStorage.getItem(LOGIN); },
        save(tokens, login = this.login) {
            localStorage.setItem(ACCESS, tokens.accessToken);
            localStorage.setItem(REFRESH, tokens.refreshToken);
            if (login) localStorage.setItem(LOGIN, login);
        },
        clear() {
            [ACCESS, REFRESH, LOGIN, GAME].forEach(key => localStorage.removeItem(key));
        },
        async auth(path, login, password) {
            const response = await fetch(path, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ login, password })
            });
            if (!response.ok) throw new Error(await message(response));
            this.save(await response.json(), login);
        },
        async refresh() {
            if (!this.refreshToken) return false;

            const response = await fetch("/auth/refresh", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ refreshToken: this.refreshToken })
            });

            if (!response.ok) return false;

            this.save(await response.json());
            return true;
        },
        async request(path, options = {}, retry = true) {
            if (!this.token) {
                location.href = "/login";
                throw new Error("Требуется авторизация");
            }

            const headers = { ...(options.headers || {}), Authorization: `Bearer ${this.token}` };
            if (options.body) headers["Content-Type"] = "application/json";

            const response = await fetch(path, { ...options, headers });

            if (response.status === 401 && retry && await this.refresh()) {
                return this.request(path, options, false);
            }

            if (response.status === 401) {
                this.clear();
                location.href = "/login";
            }

            return response;
        }
    };

    // Sidebar session.
    if ($("session-user")) {
        const loggedIn = Boolean(api.token);
        $("session-user").textContent = api.login || "Гость";
        $("session-led").classList.toggle("online", loggedIn);
        $("session-led").title = loggedIn ? "Авторизован" : "Не авторизован";
        $("login-link").classList.toggle("hidden", loggedIn);
        $("logout-button").classList.toggle("hidden", !loggedIn);
        $("logout-button").onclick = () => { api.clear(); location.href = "/login"; };
    }

    // Login / signup.
    $("login-form")?.addEventListener("submit", async event => {
        event.preventDefault();
        $("login-error").textContent = "";
        try {
            await api.auth("/auth/login", $("login").value.trim(), $("password").value);
            location.href = "/play";
        } catch (e) { $("login-error").textContent = e.message; }
    });

    $("signup-form")?.addEventListener("submit", async event => {
        event.preventDefault();
        $("signup-error").textContent = "";
        try {
            await api.auth("/signUp", $("signup-login").value.trim(), $("signup-password").value);
            location.href = "/play";
        } catch (e) { $("signup-error").textContent = e.message; }
    });

    // Leaderboard.
    async function loadLeaderboard() {
        if (!$("leaderboard-body")) return;
        try {
            const response = await api.request("/scoreboard?limit=10");
            if (!response.ok) throw new Error(await message(response));
            const rows = await response.json();
            $("leaderboard-body").innerHTML = rows.map((row, i) => {
                const rate = `${(Number(row.winRate || 0) * 100).toFixed(1)}%`;
                return `<tr class="${row.userLogin === api.login ? "me" : ""}"><td>${i + 1}</td><td>${esc(row.userLogin)}</td><td>${rate}</td></tr>`;
            }).join("") || `<tr><td colspan="3">Нет данных</td></tr>`;
        } catch (e) {
            $("leaderboard-error").textContent = e.message;
            $("leaderboard-error").classList.remove("hidden");
        }
    }
    $("reload-leaderboard")?.addEventListener("click", loadLeaderboard);
    loadLeaderboard();

    // Shared board rendering.
    function mark(value) {
        return value === 1 ? "X" : value === 2 ? "O" : "";
    }

    function side(value, myMark) {
        if (!value) return "";
        return value === myMark ? "mine" : "opponent";
    }

    function paintCell(element, value, myMark) {
        element.textContent = mark(value);
        element.classList.remove("mine", "opponent");
        const cellSide = side(value, myMark);
        if (cellSide) element.classList.add(cellSide);
    }

    function boardHTML(board, myMark) {
        return (board || []).flat().map(value =>
            `<span class="tic-cell ${side(value, myMark)}">${mark(value)}</span>`
        ).join("");
    }

    // History.
    async function loadHistory() {
        if (!$("history-list")) return;
        try {
            const userResponse = await api.request("/user");
            const gamesResponse = await api.request("/userGames");
            if (!userResponse.ok) throw new Error(await message(userResponse));
            if (!gamesResponse.ok) throw new Error(await message(gamesResponse));
            const { userId } = await userResponse.json();
            const games = await gamesResponse.json();

            $("history-list").innerHTML = games.map(game => {
                const myMark = game.ownerId === userId ? 1 : 2;
                const result = String(game.result || "").toUpperCase();
                const resultClass = result.toLowerCase();
                const date = game.createdAt ? new Date(game.createdAt).toLocaleString("ru-RU") : "—";
                return `
                    <article class="history-item">
                        <div class="tic-board mini-board">${boardHTML(game.board, myMark)}</div>
                        <div>
                            <strong>${esc(game.gameType)} · #${esc(game.id.slice(0,8))}</strong>
                            <small>${esc(date)}</small>
                        </div>
                        <span class="result ${resultClass}">${esc(result || "—")}</span>
                    </article>
                `;
            }).join("") || `<div class="card">История пуста.</div>`;
        } catch (e) {
            $("history-error").textContent = e.message;
            $("history-error").classList.remove("hidden");
        }
    }
    $("reload-history")?.addEventListener("click", loadHistory);
    loadHistory();

    // Game page.
    if ($("game-board")) {
        const WAITING = "Waiting for player connection";
        const state = { userId: null, game: null };
        const cells = [...document.querySelectorAll(".board-cell")];

        async function getUser() {
            const response = await api.request("/user");
            if (!response.ok) throw new Error(await message(response));
            state.userId = (await response.json()).userId;
        }

        function finished(game) {
            return ["Player won", "Bot won", "Draw"].includes(game?.status);
        }

        function renderGame() {
            const game = state.game;
            $("close-game").classList.toggle("hidden", !game);
            $("game-title").textContent = game ? (game.gameType === "PVE" ? "Игра против бота" : "PvP") : "Нет активной игры";
            $("game-id").textContent = game ? `#${game.id.slice(0,8)}` : "—";
            $("game-type").textContent = game?.gameType || "—";
            $("game-role").textContent = game ? (game.ownerId === state.userId ? "X" : "O") : "—";
            $("game-status").textContent = game?.status || "Создай игру или выбери существующую.";

            $("create-pve").classList.toggle("primary", game?.gameType === "PVE");
            $("create-pvp").classList.toggle("primary", game?.gameType === "PVP");

            const myMark = game ? (game.ownerId === state.userId ? 1 : 2) : 0;

            cells.forEach(cell => {
                const r = Number(cell.dataset.row), c = Number(cell.dataset.col);
                const value = game?.board?.[r]?.[c] || 0;
                paintCell(cell, value, myMark);
                cell.disabled = !game || value !== 0 || finished(game) || game.status === WAITING || game.moveUserId !== state.userId;
            });
        }

        function saveGame(game) {
            state.game = game;
            localStorage.setItem(GAME, game.id);
            renderGame();
        }

        async function create(type) {
            const response = await api.request("/game", { method: "POST", body: JSON.stringify({ gameType: type }) });
            if (!response.ok) throw new Error(await message(response));
            saveGame(await response.json());
            loadGames();
        }

        async function open(id, join = false) {
            const response = await api.request(`/game/${id}${join ? "/join" : ""}`, { method: join ? "POST" : "GET" });
            if (!response.ok) throw new Error(await message(response));
            saveGame(await response.json());
            loadGames();
        }

        async function loadGames() {
            try {
                const response = await api.request("/games");
                if (!response.ok) throw new Error(await message(response));
                const games = await response.json();
                $("games-body").innerHTML = games.map(game => {
                    const join = game.status === WAITING;
                    return `<tr><td>#${esc(game.id.slice(0,8))}</td><td>${esc(game.gameType)}</td><td>${esc(game.opponentLogin || "—")}</td><td>${esc(game.status)}</td><td><button class="link-button game-open" data-id="${esc(game.id)}" data-join="${join}">${join ? "Join" : "Open"}</button></td></tr>`;
                }).join("") || `<tr><td colspan="5">Нет доступных игр</td></tr>`;
                document.querySelectorAll(".game-open").forEach(button => {
                    button.onclick = () => open(button.dataset.id, button.dataset.join === "true").catch(showError);
                });
            } catch (e) { showError(e); }
        }

        async function move(cell) {
            if (!state.game) return;
            const board = state.game.board.map(row => [...row]);
            const r = Number(cell.dataset.row), c = Number(cell.dataset.col);
            if (board[r][c]) return;
            board[r][c] = state.game.ownerId === state.userId ? 1 : 2;
            const response = await api.request(`/game/${state.game.id}`, { method: "POST", body: JSON.stringify({ board }) });
            if (!response.ok) throw new Error(await message(response));
            saveGame(await response.json());
        }

        async function refreshGame() {
            if (!state.game || finished(state.game)) return;
            const response = await api.request(`/game/${state.game.id}`);
            if (response.ok) saveGame(await response.json());
        }

        function showError(e) {
            $("game-error").textContent = e.message || e;
            $("game-error").classList.remove("hidden");
        }

        cells.forEach(cell => cell.onclick = () => move(cell).catch(showError));
        $("create-pve").onclick = () => create("PVE").catch(showError);
        $("create-pvp").onclick = () => create("PVP").catch(showError);
        $("reload-games").onclick = loadGames;
        $("close-game").onclick = () => { state.game = null; localStorage.removeItem(GAME); renderGame(); };

        (async () => {
            try {
                await getUser();
                const id = localStorage.getItem(GAME);
                if (id) await open(id);
                else renderGame();
                await loadGames();
                setInterval(refreshGame, 2000);
            } catch (e) { showError(e); }
        })();
    }
})();
