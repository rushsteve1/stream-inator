package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"
)

type State struct {
	// Match state
	Match Match

	// Comms state
	Comms []Commentator

	// Lower third state
	LowerThird string
	Timeout    time.Time

	// Bracket state
	Bracket Bracket
}

var GlobalState State

func (s *State) Reset() {
	s.Match = Match{}
	s.LowerThird = ""
	s.Timeout = time.Time{}
	s.Bracket = Bracket{}
}

func (s *State) Save() {
	file, err := os.Create(GlobalConfig.OutFile)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
	defer file.Close()

	dec := json.NewEncoder(file)
	dec.SetIndent("", "  ")
	err = dec.Encode(s)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	slog.Info("State saved")
}

func (s *State) Load() {
	file, err := os.Open(GlobalConfig.OutFile)
	if err != nil {
		slog.Warn("Couldn't load state", "error", err.Error())
		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(s)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	slog.Info("State loaded")
}

type Player struct {
	Name string
	Team string
}

func (p Player) String() string {
	if p.Team == "" {
		return p.Name
	}
	return fmt.Sprintf("[%s] %s", p.Team, p.Name)
}

type Match struct {
	Round string

	P1      Player
	P1Score int64

	P2      Player
	P2Score int64
}

func (m *Match) Swap() {
	m.P1, m.P2 = m.P2, m.P1
	m.P1Score, m.P2Score = m.P2Score, m.P1Score
}

type Commentator struct {
	Player
	Social   string
	Pronouns string
}

type Bracket struct {
	Matches map[string]Match
	ByRound map[string][]Match
}
