(() => {
    if (!API.requireAuth()) return;
    const body = document.getElementById("leaderboard-body");
    const loading = document.getElementById("leaderboard-loading");
    const errorBox = document.getElementById("leaderboard-error");

    async function load() {
        loading.classList.remove("hidden");
        errorBox.classList.add("hidden");
        body.innerHTML = "";
        try {
            const response = await API.request("/scoreboard?limit=10");
            if (!response.ok) throw new Error(await API.errorText(response));
            const rows = await response.json();
            if (!rows || rows.length === 0) {
                body.innerHTML = `<tr><td colspan="3" class="empty-cell">Пока нет данных для рейтинга.</td></tr>`;
                return;
            }
            rows.forEach((item, index) => {
                const login = item.UserLogin ?? item.userLogin ?? "unknown";
                const raw = Number(item.Winrate ?? item.winRate ?? 0);
                const percent = Number.isFinite(raw) ? `${(raw * 100).toFixed(1)}%` : "—";
                const tr = document.createElement("tr");
                if (API.loginName && login === API.loginName) tr.classList.add("is-me");
                tr.innerHTML = `<td>${String(index + 1).padStart(2, "0")}</td><td>${escapeHtml(login)}</td><td>${percent}</td>`;
                body.appendChild(tr);
            });
        } catch (e) {
            errorBox.textContent = e.message;
            errorBox.classList.remove("hidden");
        } finally {
            loading.classList.add("hidden");
        }
    }

    function escapeHtml(value) {
        return String(value).replace(/[&<>'"]/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;","'":"&#39;",'"':"&quot;"}[c]));
    }

    document.getElementById("reload-leaderboard")?.addEventListener("click", load);
    load();
})();
