(() => {
    if (!API.requireAuth()) return;

    const list = document.getElementById("history-list");
    const loading = document.getElementById("history-loading");
    const errorBox = document.getElementById("history-error");
    const count = document.getElementById("history-count");

    async function load() {
        loading.classList.remove("hidden");
        errorBox.classList.add("hidden");
        list.innerHTML = "";

        try {
            const response = await API.request("/userGames");
            if (!response.ok) throw new Error(await API.errorText(response));

            const games = await response.json();
            count.textContent = `${games?.length || 0} MATCHES`;

            if (!games || games.length === 0) {
                list.innerHTML = `<article class="panel empty-panel">Завершённых игр пока нет.</article>`;
                return;
            }

            games
                .sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt))
                .forEach(game => {
                    const date = game.createdAt
                        ? new Date(game.createdAt).toLocaleString("ru-RU")
                        : "—";

                    const row = document.createElement("article");
                    row.className = "match-row";
                    row.innerHTML = `
                        <span class="match-id">#${escapeHtml(game.id.slice(0, 8))}</span>
                        <div>
                            <strong>${escapeHtml(game.gameType)}</strong>
                            <span>${escapeHtml(date)}</span>
                        </div>
                        <div class="mini-board">${miniBoard(game.board)}</div>
                        <span class="result ${game.status === "Player won" ? "result--win" : ""}">${escapeHtml(game.status)}</span>`;

                    list.appendChild(row);
                });
        } catch (e) {
            errorBox.textContent = e.message;
            errorBox.classList.remove("hidden");
        } finally {
            loading.classList.add("hidden");
        }
    }

    function miniBoard(board) {
        if (!Array.isArray(board)) return "";

        return board
            .flat()
            .map(value => `<span>${value === 1 ? "X" : value === 2 ? "O" : "·"}</span>`)
            .join("");
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

    document.getElementById("reload-history")?.addEventListener("click", load);
    load();
})();
