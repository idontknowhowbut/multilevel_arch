(() => {
    if (!API.requireAuth()) return;

    const state = {
        userId: null,
        game: null,
        ownerId: null,
        waiting: false,
        games: [],
        gameTimer: null,
        gamesTimer: null,
    };

    const cells = [...document.querySelectorAll(".board-cell")];
    const gameTitle = document.getElementById("game-title");
    const gameId = document.getElementById("game-id");
    const gameType = document.getElementById("game-type");
    const gameRole = document.getElementById("game-role");
    const gameStatus = document.getElementById("game-status");
    const gameError = document.getElementById("game-error");
    const gamesBody = document.getElementById("games-body");
    const gamesLoading = document.getElementById("games-loading");
    const gamesError = document.getElementById("games-error");
    const closeGameButton = document.getElementById("close-game");

    async function init() {
        try {
            const response = await API.request("/user");
            if (!response.ok) throw new Error(await API.errorText(response));
            const user = await response.json();
            state.userId = user.userId;
            await restoreGame();
            await loadGames();
            startPolling();
        } catch (e) {
            showGameError(e.message);
        }
    }

    function normalizeGame(raw) {
        if (!raw) return null;
        return {
            id: raw.id ?? raw.Id,
            board: raw.board ?? raw.Board ?? [[0,0,0],[0,0,0],[0,0,0]],
            type: raw.gameType ?? raw.Type ?? "—",
            status: raw.status ?? raw.Status ?? "",
            moveUserId: raw.moveUserId ?? raw.MovePlayerId ?? "",
            ownerId: raw.ownerId ?? raw.OwnerId ?? null,
            createdAt: raw.createdAt ?? raw.CreatedAt ?? null,
        };
    }

    async function createGame(type) {
        clearGameError();
        try {
            const response = await API.request("/game", {
                method: "POST",
                body: JSON.stringify({ gameType: type }),
            });
            if (!response.ok) throw new Error(await API.errorText(response));
            state.game = normalizeGame(await response.json());
            state.ownerId = state.userId;
            state.waiting = type === "PVP";
            persistGame();
            renderGame();
            await loadGames();
        } catch (e) {
            showGameError(e.message);
        }
    }

    async function loadGames() {
        gamesLoading.classList.remove("hidden");
        gamesError.classList.add("hidden");
        try {
            const response = await API.request("/games");
            if (!response.ok) throw new Error(await API.errorText(response));
            const raw = await response.json();
            state.games = (raw || []).map(normalizeGame);
            renderGames();

            if (state.waiting && state.game) {
                const active = state.games.find(g => g.id === state.game.id && g.status === "Waiting for player move");
                if (active) {
                    state.waiting = false;
                    state.ownerId = active.ownerId || state.ownerId;
                    localStorage.setItem("ttt.currentGameWaiting", "false");
                    await openGame(active.id, state.ownerId);
                }
            }
        } catch (e) {
            gamesError.textContent = e.message;
            gamesError.classList.remove("hidden");
        } finally {
            gamesLoading.classList.add("hidden");
        }
    }

    function renderGames() {
        gamesBody.innerHTML = "";
        if (!state.games.length) {
            gamesBody.innerHTML = `<tr><td colspan="5" class="empty-cell">Доступных или активных игр пока нет.</td></tr>`;
            return;
        }
        state.games.forEach(game => {
            const waitingForPlayer = game.status === "Waiting for player connection";
            const actionLabel = waitingForPlayer ? "JOIN" : "OPEN";
            const owner = game.ownerId ? game.ownerId.slice(0, 8) : "—";
            const tr = document.createElement("tr");
            if (state.game?.id === game.id) tr.classList.add("is-current-game");
            tr.innerHTML = `
                <td>#${escapeHtml(game.id.slice(0,8))}</td>
                <td>${escapeHtml(game.type)}</td>
                <td>${escapeHtml(game.status || "—")}</td>
                <td>${escapeHtml(owner)}</td>
                <td><button class="table-action" data-game="${escapeHtml(game.id)}" data-owner="${escapeHtml(game.ownerId || "")}" data-action="${waitingForPlayer ? "join" : "open"}">${actionLabel}</button></td>`;
            gamesBody.appendChild(tr);
        });
        gamesBody.querySelectorAll(".table-action").forEach(button => {
            button.addEventListener("click", async () => {
                const id = button.dataset.game;
                const owner = button.dataset.owner;
                if (button.dataset.action === "join") await joinGame(id, owner);
                else await openGame(id, owner);
            });
        });
    }

    async function joinGame(id, ownerId) {
        clearGameError();
        try {
            const response = await API.request(`/game/${id}/join`, { method: "POST" });
            if (!response.ok) throw new Error(await API.errorText(response));
            state.ownerId = ownerId;
            state.waiting = false;
            await openGame(id, ownerId);
            await loadGames();
        } catch (e) {
            showGameError(e.message);
        }
    }

    async function openGame(id, ownerId = null) {
        clearGameError();
        try {
            const response = await API.request(`/game/${id}`);
            if (!response.ok) throw new Error(await API.errorText(response));
            state.game = normalizeGame(await response.json());
            if (ownerId) state.ownerId = ownerId;
            state.waiting = false;
            persistGame();
            renderGame();
        } catch (e) {
            showGameError(e.message);
        }
    }

    async function restoreGame() {
        const id = localStorage.getItem("ttt.currentGameId");
        if (!id) {
            renderGame();
            return;
        }
        state.ownerId = localStorage.getItem("ttt.currentGameOwnerId");
        state.waiting = localStorage.getItem("ttt.currentGameWaiting") === "true";
        try {
            const response = await API.request(`/game/${id}`);
            if (!response.ok) throw new Error("Игра больше недоступна");
            state.game = normalizeGame(await response.json());
            renderGame();
        } catch (_) {
            clearCurrentGame();
        }
    }

    function persistGame() {
        if (!state.game) return;
        localStorage.setItem("ttt.currentGameId", state.game.id);
        if (state.ownerId) localStorage.setItem("ttt.currentGameOwnerId", state.ownerId);
        localStorage.setItem("ttt.currentGameWaiting", String(state.waiting));
    }

    function clearCurrentGame() {
        state.game = null;
        state.ownerId = null;
        state.waiting = false;
        localStorage.removeItem("ttt.currentGameId");
        localStorage.removeItem("ttt.currentGameOwnerId");
        localStorage.removeItem("ttt.currentGameWaiting");
        renderGame();
    }

    function renderGame() {
        const game = state.game;
        if (!game) {
            gameTitle.textContent = "Нет активной игры";
            gameId.textContent = "—";
            gameType.textContent = "—";
            gameRole.textContent = "—";
            gameStatus.textContent = "Создай игру или выбери доступную.";
            closeGameButton.classList.add("hidden");
            cells.forEach(cell => { cell.textContent = ""; cell.disabled = true; cell.className = "board-cell"; });
            return;
        }

        closeGameButton.classList.remove("hidden");
        gameTitle.textContent = game.type === "PVE" ? "Игра против бота" : "PvP матч";
        gameId.textContent = `#${game.id.slice(0, 8)}`;
        gameType.textContent = game.type;
        const myMark = state.ownerId === state.userId || game.type === "PVE" ? 1 : 2;
        gameRole.textContent = myMark === 1 ? "X" : "O";

        const end = getGameEnd(game.board);
        let statusText;
        if (state.waiting) statusText = "Ожидаем второго игрока...";
        else if (end.over && end.winner === 0) statusText = "Ничья.";
        else if (end.over) statusText = end.winner === myMark ? "Победа." : (game.type === "PVE" ? "Бот победил." : "Соперник победил.");
        else if (game.moveUserId === state.userId) statusText = "Ваш ход.";
        else statusText = "Ход соперника.";
        gameStatus.textContent = statusText;

        cells.forEach(cell => {
            const r = Number(cell.dataset.row), c = Number(cell.dataset.col);
            const value = game.board?.[r]?.[c] ?? 0;
            cell.textContent = value === 1 ? "X" : value === 2 ? "O" : "";
            cell.className = `board-cell ${value === 1 ? "board-cell--x" : value === 2 ? "board-cell--o" : ""}`;
            cell.disabled = value !== 0 || end.over || state.waiting || game.moveUserId !== state.userId;
        });
    }

    async function makeMove(row, col) {
        if (!state.game || state.waiting) return;
        clearGameError();
        const board = state.game.board.map(r => [...r]);
        if (board[row][col] !== 0) return;
        board[row][col] = 1;
        try {
            const response = await API.request(`/game/${state.game.id}`, {
                method: "POST",
                body: JSON.stringify({ board }),
            });
            if (!response.ok) throw new Error(await API.errorText(response));
            const updated = normalizeGame(await response.json());
            updated.ownerId = state.ownerId;
            state.game = updated;
            persistGame();
            renderGame();
            if (getGameEnd(updated.board).over) await loadGames();
        } catch (e) {
            showGameError(e.message);
            await refreshCurrentGame();
        }
    }

    async function refreshCurrentGame() {
        if (!state.game || state.waiting) return;
        try {
            const response = await API.request(`/game/${state.game.id}`);
            if (!response.ok) return;
            const updated = normalizeGame(await response.json());
            updated.ownerId = state.ownerId;
            state.game = updated;
            renderGame();
        } catch (_) {}
    }

    function getGameEnd(board) {
        if (!board) return { over: false, winner: 0 };
        const lines = [
            [[0,0],[0,1],[0,2]], [[1,0],[1,1],[1,2]], [[2,0],[2,1],[2,2]],
            [[0,0],[1,0],[2,0]], [[0,1],[1,1],[2,1]], [[0,2],[1,2],[2,2]],
            [[0,0],[1,1],[2,2]], [[2,0],[1,1],[0,2]],
        ];
        for (const line of lines) {
            const [a,b,c] = line.map(([r,col]) => board[r][col]);
            if (a !== 0 && a === b && b === c) return { over: true, winner: a };
        }
        return { over: board.flat().every(v => v !== 0), winner: 0 };
    }

    function startPolling() {
        state.gamesTimer = setInterval(loadGames, 4000);
        state.gameTimer = setInterval(() => {
            if (state.game && !state.waiting && state.game.type === "PVP" && !getGameEnd(state.game.board).over) {
                refreshCurrentGame();
            }
        }, 1800);
    }

    function showGameError(message) {
        gameError.textContent = message;
        gameError.classList.remove("hidden");
    }
    function clearGameError() { gameError.classList.add("hidden"); gameError.textContent = ""; }
    function escapeHtml(value) {
        return String(value).replace(/[&<>'"]/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;","'":"&#39;",'"':"&quot;"}[c]));
    }

    cells.forEach(cell => cell.addEventListener("click", () => makeMove(Number(cell.dataset.row), Number(cell.dataset.col))));
    document.getElementById("create-pve")?.addEventListener("click", () => createGame("PVE"));
    document.getElementById("create-pvp")?.addEventListener("click", () => createGame("PVP"));
    document.getElementById("reload-games")?.addEventListener("click", loadGames);
    closeGameButton?.addEventListener("click", clearCurrentGame);
    window.addEventListener("beforeunload", () => { clearInterval(state.gameTimer); clearInterval(state.gamesTimer); });
    init();
})();
