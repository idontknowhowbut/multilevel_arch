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

            games.sort((a, b) => new Date(b.CreatedAt ?? b.createdAt) - new Date(a.CreatedAt ?? a.createdAt));
            games.forEach(game => {
                const id = game.Id ?? game.id ?? "";
                const type = game.Type ?? game.gameType ?? "—";
                const status = game.Status ?? game.status ?? "—";
                const dateValue = game.CreatedAt ?? game.createdAt;
                const date = dateValue ? new Date(dateValue).toLocaleString("ru-RU") : "—";
                const board = game.Board ?? game.board;
                const row = document.createElement("article");
                row.className = "match-row";
                row.innerHTML = `
                    <span class="match-id">#${escapeHtml(id.slice(0, 8))}</span>
                    <div><strong>${escapeHtml(type)}</strong><span>${escapeHtml(date)}</span></div>
                    <div class="mini-board">${miniBoard(board)}</div>
                    <span class="result ${status === "Player won" ? "result--win" : ""}">${escapeHtml(status)}</span>`;
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
        return board.flat().map(v => `<span>${v === 1 ? "X" : v === 2 ? "O" : "·"}</span>`).join("");
    }
    function escapeHtml(value) {
        return String(value).replace(/[&<>'"]/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;","'":"&#39;",'"':"&quot;"}[c]));
    }

    document.getElementById("reload-history")?.addEventListener("click", load);
    load();
})();
