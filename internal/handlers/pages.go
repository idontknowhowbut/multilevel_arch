package handlers

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

const sourceUrl = "https://github.com/idontknowhowbut/multilevel_arch"

type PageData struct {
	Title     string
	Active    string
	Year      int
	GitHubURL string
}

func renderPage(
	w http.ResponseWriter,
	page string,
	data PageData,
) {
	files := []string{
		"templates/layout.html",
		"templates/components/sidebar.html",
		"templates/components/footer.html",
		filepath.Join("templates/pages", page+".html"),
	}

	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "home", PageData{
		Title:  "Tic Tac Toe",
		Active: "home",
		Year:   time.Now().Year(),
	})
}

func (h *Handler) PlayPage(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "play", PageData{
		Title:  "Tic Tac Toe - Play",
		Active: "play",
		Year:   time.Now().Year(),
	})
}

func (h *Handler) LeaderboardPage(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "leaderboard", PageData{
		Title:  "Tic Tac Toe - Leaderboard",
		Active: "leaderboard",
		Year:   time.Now().Year(),
	})
}

func (h *Handler) HistoryPage(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "history", PageData{
		Title:  "Tic Tac Toe - History",
		Active: "history",
		Year:   time.Now().Year(),
	})
}

func (h *Handler) AboutPage(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "about", PageData{
		Title:     "Tic Tac Toe - About",
		Active:    "about",
		Year:      time.Now().Year(),
		GitHubURL: sourceUrl,
	})
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "login", PageData{
		Title:  "Tic Tac Toe — Вход",
		Active: "",
		Year:   time.Now().Year(),
	})
}
