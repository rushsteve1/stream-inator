package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

//go:embed static/*
var staticFS embed.FS

func registerHandlers() {
	http.HandleFunc("GET /state.json", stateJsonGET)

	http.HandleFunc("GET /", indexGET)

	http.Handle("GET /static/", http.FileServerFS(staticFS))

	http.HandleFunc("GET /match", matchGET)
	http.HandleFunc("POST /match", matchPOST)
	http.HandleFunc("DELETE /match", matchDELETE)
	http.HandleFunc("PATCH /match", matchPATCH)

	http.HandleFunc("GET /config", configGET)
	http.HandleFunc("POST /config", configPOST)
	http.HandleFunc("DELETE /config", configDELETE)
	http.HandleFunc("GET /config.json", configJsonGET)

	http.HandleFunc("GET /comms", commsGET)
	http.HandleFunc("POST /comms", commsPOST)
	http.HandleFunc("DELETE /comms", commsDELETE)

	http.HandleFunc("GET /lower", lowerGET)

	http.HandleFunc("GET /bracket", bracketGET)
}

func stateJsonGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(GlobalState)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Error encoding state", http.StatusInternalServerError)
	}
}

func indexGET(w http.ResponseWriter, r *http.Request) {
	type indexData struct {
		OutFile string
		LocalIP string
	}

	data := indexData{GlobalConfig.OutFile, localIPAddress()}

	err := tmpl("index").Execute(w, data)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
	}
}

func matchGET(w http.ResponseWriter, r *http.Request) {

	type matchData struct {
		Match   Match
		Players []string
		Teams   []string
		Rounds  []string
	}

	data := matchData{
		GlobalState.Match,
		GlobalConfig.Players,
		GlobalConfig.Teams,
		GlobalConfig.Rounds,
	}

	err := tmpl("match").Execute(w, data)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
	}
}

func matchPOST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}

	GlobalState.Match.Round = r.Form.Get("round")
	GlobalState.Match.P1.Name = r.Form.Get("p1name")
	GlobalState.Match.P1.Team = r.Form.Get("p1team")
	GlobalState.Match.P2Score, err = strconv.ParseInt(r.Form.Get("p1score"), 10, 64)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Error parsing p1 score", http.StatusBadRequest)
	}

	GlobalState.Match.P2.Name = r.Form.Get("p2name")
	GlobalState.Match.P2.Team = r.Form.Get("p2team")
	GlobalState.Match.P2Score, err = strconv.ParseInt(r.Form.Get("p2score"), 10, 64)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Error parsing p2 score", http.StatusBadRequest)
	}

	GlobalState.Match.Round = r.Form.Get("round")

	go GlobalState.Save()

	http.Redirect(w, r, "/match", http.StatusSeeOther)
}

func matchDELETE(w http.ResponseWriter, r *http.Request) {
	GlobalState.Match = Match{}

	go GlobalState.Save()

	http.Redirect(w, r, "/match", http.StatusSeeOther)
}

func matchPATCH(w http.ResponseWriter, r *http.Request) {
	GlobalState.Match.Swap()

	go GlobalState.Save()

	http.Redirect(w, r, "/match", http.StatusSeeOther)
}

func configGET(w http.ResponseWriter, r *http.Request) {
	type configData struct {
		Path   string
		Config string
	}

	cfg, err := json.MarshalIndent(GlobalConfig, "", "  ")
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Error reading config", http.StatusInternalServerError)
	}
	cfgstr := strings.TrimSpace(string(cfg))

	err = tmpl("config").Execute(w, configData{GlobalConfig.Path(), cfgstr})
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
	}
}

func configPOST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	err = GlobalConfig.Update(string(r.Form.Get("config")))
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/config", http.StatusSeeOther)
}

func configDELETE(w http.ResponseWriter, r *http.Request) {
	GlobalConfig.Reset()

	http.Redirect(w, r, "/config", http.StatusSeeOther)
}

func configJsonGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(GlobalConfig)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Error encoding config", http.StatusInternalServerError)
	}
}

func commsGET(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	type commsData struct {
		Comms []Commentator
		Teams []string
		Names []string
	}

	comms := make([]Commentator, len(GlobalState.Comms))
	copy(comms, GlobalState.Comms)

	commLen := 0
	if r.Form.Has("l") {
		commLen, err = strconv.Atoi(r.Form.Get("l"))
		if err != nil {
			slog.ErrorContext(r.Context(), err.Error())
			http.Error(w, "Error parsing length", http.StatusBadRequest)
			return
		}
	} else {
		http.Redirect(w, r, fmt.Sprintf("/comms?l=%d", max(len(comms), 1)), http.StatusSeeOther)
	}

	if commLen < len(comms) {
		commLen = len(comms)
		http.Redirect(w, r, fmt.Sprintf("/comms?l=%d", commLen), http.StatusSeeOther)
		return
	}

	commLen = max(commLen, 1)

	if commLen > len(comms) {
		for i := len(comms); i < commLen; i++ {
			comms = append(comms, Commentator{})
		}
	}

	data := commsData{comms, GlobalConfig.Teams, GlobalConfig.Commentators}

	err = tmpl("comms").Execute(w, data)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
	}
}

func commsPOST(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Error parsing form", http.StatusBadRequest)
	}

	GlobalState.Comms = make([]Commentator, len(r.Form["name"]))

	for i, name := range r.Form["name"] {
		if len(name) == 0 {
			continue
		}

		GlobalState.Comms[i].Name = name
		GlobalState.Comms[i].Team = r.Form["team"][i]
		GlobalState.Comms[i].Social = r.Form["social"][i]
		GlobalState.Comms[i].Pronouns = r.Form["pronouns"][i]
	}

	go GlobalState.Save()

	http.Redirect(w, r, "/comms", http.StatusSeeOther)
}

func commsDELETE(w http.ResponseWriter, r *http.Request) {
	GlobalState.Comms = nil

	go GlobalState.Save()

	http.Redirect(w, r, "/comms", http.StatusSeeOther)
}

func lowerGET(w http.ResponseWriter, r *http.Request) {
	err := tmpl("lower").Execute(w, GlobalState)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Error rendering lower third", http.StatusInternalServerError)
	}
}

func bracketGET(w http.ResponseWriter, r *http.Request) {
	err := tmpl("bracket").Execute(w, nil)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
	}
}
