package pastgames

import (
	"html/template"
	"net/http"
	"strconv"
)

const (
	PastGameTemplate = "internal/pastgames/templates/pastgame.html"
	BaseTemplate     = "internal/common/templates/base.html"
)

var pastGameRepo PastGameRepository = NewInMemoryPastGameRepository()

func GetPastGameHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("game-id")
	if idStr == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	pastGame, err := pastGameRepo.GetPastGameByID(id)
	if err != nil {
		if _, ok := err.(ErrPastGameNotFound); ok {
			http.Error(w, "Past game not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles(PastGameTemplate, BaseTemplate))
	if err := tmpl.ExecuteTemplate(w, "base", pastGame); err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}
