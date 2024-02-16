package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
)

type Config struct {
	Port    int
	OutFile string
	Teams   []string
	Players []string
	Rounds  []string

	path string
}

var GlobalConfig Config

func (c *Config) Load(path string) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.Default()
			c.CreateConfig(path)
			return
		}
		slog.Error(err.Error())
		panic(err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(c)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	c.path = path
}

func (c *Config) Path() string {
	return c.path
}

func (c *Config) Default() {
	c.Port = 9090
	c.Rounds = []string{"Round 1", "Round 2", "Round 3"}

	cwd, err := os.Getwd()
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
	c.OutFile = filepath.Join(cwd, "stream.json")
}

func (c *Config) DefaultPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	return filepath.Join(dir, "stream-inator", "config.json")
}

func (c *Config) CreateConfig(path string) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	file, err := os.Create(path)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(c)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	c.path = path
}

func (c *Config) Update(newConfig string) {
	err := json.Unmarshal([]byte(newConfig), c)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}

	file, err := os.Create(c.Path())
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(c)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
}

func (c *Config) Reset() {
	c.Default()
	c.CreateConfig(c.Path())
}
