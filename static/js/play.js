(() => {
    if (!API.requireAuth()) return;

    const STATUS_WAITING_PLAYER = "Waiting for player connection";

    const state = {
        userId: null,
        game: null,
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

    async function createGame(type) {
        clearGameError();

        try {
            const response = await API.request("/game", {
                method: "POST",
                body: JSON.stringify({ gameType: type }),
            });

            if (!response.ok) throw new Error(await API.errorText(response));

            state.game = await response.json();
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

            state.games = await response.json() || [];
            renderGames();
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
            const waitingForPlayer = game.status === STATUS_WAITING_PLAYER;
            const actionLabel = waitingForPlayer ? "JOIN" : "OPEN";
            const owner = game.ownerId ? game.ownerId.slice(0, 8) : "—";
            const tr = document.createElement("tr");

            if (state.game?.id === game.id) {
                tr.classList.add("is-current-game");
            }

            tr.innerHTML = `
                <td>#${escapeHtml(game.id.slice(0, 8))}</td>
                <td>${escapeHtml(game.gameType)}</td>
                <td>${escapeHtml(game.status)}</td>
                <td>${escapeHtml(owner)}</td>
                <td>
                    <button
                        class="table-action"
                        data-game="${escapeHtml(game.id)}"
                        data-action="${waitingForPlayer ? "join" : "open"}"
                        type="button"
                    >${actionLabel}</button>
                </td>`;

            gamesBody.appendChild(tr);
        });

        gamesBody.querySelectorAll(".table-action").forEach(button => {
            button.addEventListener("click", async () => {
                const id = button.dataset.game;

                if (button.dataset.action === "join") {
                    await joinGame(id);
                } else {
                    await openGame(id);
                }
            });
        });
    }

    async function joinGame(id) {
        clearGameError();

        try {
            const response = await API.request(`/game/${id}/join`, {
                method: "POST",
            });

            if (!response.ok) throw new Error(await API.errorText(response));

            state.game = await response.json();
            persistGame();
            renderGame();
            await loadGames();
        } catch (e) {
            showGameError(e.message);
        }
    }

    async function openGame(id) {
        clearGameError();

        try {
            const response = await API.request(`/game/${id}`);
            if (!response.ok) throw new Error(await API.errorText(response));

            state.game = await response.json();
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

        try {
            const response = await API.request(`/game/${id}`);
            if (!response.ok) throw new Error("Игра больше недоступна");

            state.game = await response.json();
            renderGame();
        } catch (_) {
            clearCurrentGame();
        }
    }

    function persistGame() {
        if (!state.game) return;
        localStorage.setItem("ttt.currentGameId", state.game.id);
    }

    function clearCurrentGame() {
        state.game = null;
        localStorage.removeItem("ttt.currentGameId");
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

            cells.forEach(cell => {
                cell.textContent = "";
                cell.disabled = true;
                cell.className = "board-cell";
            });

            return;
        }

        closeGameButton.classList.remove("hidden");
        gameTitle.textContent = game.gameType === "PVE" ? "Игра против бота" : "PvP матч";
        gameId.textContent = `#${game.id.slice(0, 8)}`;
        gameType.textContent = game.gameType;

        const myMark = game.ownerId === state.userId ? 1 : 2;
        gameRole.textContent = myMark === 1 ? "X" : "O";

        const waitingForPlayer = game.status === STATUS_WAITING_PLAYER;
        const end = getGameEnd(game.board);

        if (waitingForPlayer) {
            gameStatus.textContent = "Ожидаем второго игрока...";
        } else if (end.over && end.winner === 0) {
            gameStatus.textContent = "Ничья.";
        } else if (end.over) {
            gameStatus.textContent = end.winner === myMark
                ? "Победа."
                : game.gameType === "PVE"
                    ? "Бот победил."
                    : "Соперник победил.";
        } else if (game.moveUserId === state.userId) {
            gameStatus.textContent = "Ваш ход.";
        } else {
            gameStatus.textContent = "Ход соперника.";
        }

        cells.forEach(cell => {
            const row = Number(cell.dataset.row);
            const col = Number(cell.dataset.col);
            const value = game.board[row][col];

            cell.textContent = value === 1 ? "X" : value === 2 ? "O" : "";
            cell.className = `board-cell ${value === 1 ? "board-cell--x" : value === 2 ? "board-cell--o" : ""}`;
            cell.disabled = (
                value !== 0 ||
                end.over ||
                waitingForPlayer ||
                game.moveUserId !== state.userId
            );
        });
    }

    async function makeMove(row, col) {
        if (!state.game || state.game.status === STATUS_WAITING_PLAYER) return;

        clearGameError();

        const board = state.game.board.map(boardRow => [...boardRow]);
        if (board[row][col] !== 0) return;

        const myMark = state.game.ownerId === state.userId ? 1 : 2;
        board[row][col] = myMark;

        try {
            const response = await API.request(`/game/${state.game.id}`, {
                method: "POST",
                body: JSON.stringify({ board }),
            });

            if (!response.ok) throw new Error(await API.errorText(response));

            state.game = await response.json();
            persistGame();
            renderGame();

            if (getGameEnd(state.game.board).over) {
                await loadGames();
            }
        } catch (e) {
            showGameError(e.message);
            await refreshCurrentGame();
        }
    }

    async function refreshCurrentGame() {
        if (!state.game) return;

        try {
            const response = await API.request(`/game/${state.game.id}`);
            if (!response.ok) return;

            state.game = await response.json();
            persistGame();
            renderGame();
        } catch (_) {
            // Polling failures are intentionally silent.
        }
    }

    function getGameEnd(board) {
        const lines = [
            [[0,0],[0,1],[0,2]],
            [[1,0],[1,1],[1,2]],
            [[2,0],[2,1],[2,2]],
            [[0,0],[1,0],[2,0]],
            [[0,1],[1,1],[2,1]],
            [[0,2],[1,2],[2,2]],
            [[0,0],[1,1],[2,2]],
            [[2,0],[1,1],[0,2]],
        ];

        for (const line of lines) {
            const [a, b, c] = line.map(([row, col]) => board[row][col]);
            if (a !== 0 && a === b && b === c) {
                return { over: true, winner: a };
            }
        }

        return {
            over: board.flat().every(value => value !== 0),
            winner: 0,
        };
    }

    function startPolling() {
        state.gamesTimer = setInterval(loadGames, 4000);
        state.gameTimer = setInterval(() => {
            if (state.game && state.game.gameType === "PVP" && !getGameEnd(state.game.board).over) {
                refreshCurrentGame();
            }
        }, 1800);
    }

    function showGameError(message) {
        gameError.textContent = message;
        gameError.classList.remove("hidden");
    }

    function clearGameError() {
        gameError.classList.add("hidden");
        gameError.textContent = "";
    }

    function escapeHtml(value) {
        return String(value).replace(/[&<>'"]/g, c => ({
            "&": "&amp;",
            "<": "&lt;",
            ">": "&gt;",
            "'": "&#39;",
            '"': "&quot;",
        }[c]));
    }

    cells.forEach(cell => {
        cell.addEventListener("click", () => {
            makeMove(Number(cell.dataset.row), Number(cell.dataset.col));
        });
    });

    document.getElementById("create-pve")?.addEventListener("click", () => createGame("PVE"));
    document.getElementById("create-pvp")?.addEventListener("click", () => createGame("PVP"));
    document.getElementById("reload-games")?.addEventListener("click", loadGames);
    closeGameButton?.addEventListener("click", clearCurrentGame);

    window.addEventListener("beforeunload", () => {
        clearInterval(state.gameTimer);
        clearInterval(state.gamesTimer);
    });

    init();
})();
