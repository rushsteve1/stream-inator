package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func registerHandlers() {
	http.HandleFunc("GET /state.json", stateJsonGET)

	http.HandleFunc("GET /", indexGET)

	http.HandleFunc("GET /match", matchGET)
	http.HandleFunc("POST /match", matchPOST)
	http.HandleFunc("DELETE /match", matchDELETE)
	http.HandleFunc("PATCH /match", matchPATCH)

	http.HandleFunc("GET /config", configGET)
	http.HandleFunc("POST /config", configPOST)
	http.HandleFunc("DELETE /config", configDELETE)
	http.HandleFunc("GET /config.json", configJsonGET)

	http.HandleFunc("GET /comms", commsGET)

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
	}

	data := indexData{GlobalConfig.OutFile}

	err := tmpl("index").Execute(w, data)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
	}
}

type matchData struct {
	Match   Match
	Players []string
	Teams   []string
	Rounds  []string
}

func matchGET(w http.ResponseWriter, r *http.Request) {
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

type configData struct {
	Path   string
	Config string
}

func configGET(w http.ResponseWriter, r *http.Request) {
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
	}

	GlobalConfig.Update(string(r.Form.Get("config")))

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
	err := tmpl("comms").Execute(w, nil)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
	}
}

func lowerGET(w http.ResponseWriter, r *http.Request) {
	err := tmpl("lower").Execute(w, nil)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
	}
}

func bracketGET(w http.ResponseWriter, r *http.Request) {
	err := tmpl("bracket").Execute(w, nil)
	if err != nil {
		slog.ErrorContext(r.Context(), err.Error())
	}
}
