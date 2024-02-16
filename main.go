package main

import (
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"fyne.io/systray"
	"github.com/pkg/browser"
)

//go:embed templates/*
var TemplatesFS embed.FS

func main() {
	slog.Info("Starting Stream-inator")

	configPath := flag.String("config", GlobalConfig.DefaultPath(), "Path to the config file")
	flag.Parse()

	GlobalConfig.Load(*configPath)

	systray.Run(onReady, onExit)
}

func onReady() {
	slog.Info("Systray ready")

	systray.SetTitle("Stream-inator")
	systray.SetTooltip("Stream-inator version 0")

	mQuit := systray.AddMenuItem("Quit", "Quit Stream-inator")
	go quitHandler(mQuit.ClickedCh)

	mBrowser := systray.AddMenuItem("Open", "Open Stream-inator in browser")
	go browserHandler(mBrowser.ClickedCh)

	mState := systray.AddMenuItem("Open State", "Open state file folder")
	go stateHandler(mState.ClickedCh)

	mConfig := systray.AddMenuItem("Open Config", "Open config file folder")
	go configHandler(mConfig.ClickedCh)

	GlobalState.Load()

	startServer()
}

func onExit() {
	slog.Warn("Systray exiting")
	os.Exit(0)
}

func startServer() {
	registerHandlers()

	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", GlobalConfig.Port),
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           SlogMiddleware{http.DefaultServeMux},
	}

	openBrowser()

	slog.Warn("Starting http server", "port", GlobalConfig.Port)

	err := srv.ListenAndServe()
	slog.Error(err.Error())
	panic(err)
}

func quitHandler(ch chan struct{}) {
	<-ch
	systray.Quit()
}

func browserHandler(ch chan struct{}) {
	for range ch {
		openBrowser()
	}
}

func stateHandler(ch chan struct{}) {
	for range ch {
		browser.OpenURL(filepath.Dir(GlobalConfig.OutFile))
	}
}

func configHandler(ch chan struct{}) {
	for range ch {
		browser.OpenURL(filepath.Dir(GlobalConfig.Path()))
	}
}

func openBrowser() {
	slog.Info("Opening browser")
	browser.OpenURL(fmt.Sprintf("http://localhost:%d", GlobalConfig.Port))
}
