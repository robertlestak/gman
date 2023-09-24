package gman

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"git.shdw.tech/shdw.tech/gman/pkg/release"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Repo struct {
	URL    string `json:"repo" yaml:"url"`
	Branch string `json:"branch" yaml:"branch"`
}

type ConfigFile struct {
	Interval        *time.Duration   `json:"interval" yaml:"interval"`
	Namespace       *string          `json:"namespace" yaml:"namespace"`
	OpenOnGetFail   *bool            `json:"open" yaml:"open"`
	NotifyOnRelease *bool            `json:"notify" yaml:"notify"`
	Pager           *string          `json:"pager" yaml:"pager"`
	Repo            *string          `json:"repo" yaml:"repo"`
	Render          *bool            `json:"render" yaml:"render"`
	TLDR            *bool            `json:"tldr" yaml:"tldr"`
	Repos           map[string]*Repo `json:"repos" yaml:"repos"`
	Web             *bool            `json:"web" yaml:"web"`
	WebAddr         *string          `json:"webAddr" yaml:"webAddr"`
	WebDir          *string          `json:"webDir" yaml:"webDir"`
}

func (g *Gman) LoadConfig() error {
	l := log.WithField("fn", "LoadConfig")
	l.Debug("loading config")
	if g.ConfigDir == "" {
		l.Error("config dir not set")
		return errors.New("config dir not set")
	}
	// ensure config dir exists
	_, err := os.Stat(g.ConfigDir)
	if os.IsNotExist(err) {
		l.Error("config dir does not exist")
		return errors.New("config dir does not exist")
	}
	// load config file
	configFile := filepath.Join(g.ConfigDir, "config.yaml")
	// if file doesn't exist, return nil
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		l.Debug("config file does not exist")
		return nil
	}
	// read the file
	b, err := os.ReadFile(configFile)
	if err != nil {
		l.WithError(err).Error("error reading config file")
		return err
	}
	// parse the file
	var config ConfigFile
	err = yaml.Unmarshal(b, &config)
	if err != nil {
		// try json
		err = json.Unmarshal(b, &config)
		if err != nil {
			l.WithError(err).Error("error parsing config file")
			return err
		}
	}
	if g.Repo != nil && g.Repo.URL != "" && config.Repos[g.Repo.URL] != nil {
		g.Repo = config.Repos[g.Repo.URL]
	}
	if g.Repo == nil || g.Repo.URL == "" && config.Repo != nil && config.Repos[*config.Repo] != nil {
		g.Repo = config.Repos[*config.Repo]
	}
	if config.OpenOnGetFail != nil {
		OpenURLOnGetFailure = *config.OpenOnGetFail
		release.OpenURLOnGetFailure = *config.OpenOnGetFail
	}
	if config.NotifyOnRelease != nil {
		g.NotifyOnNewRelease = *config.NotifyOnRelease
	}
	if config.Interval != nil {
		g.UpdateInterval = *config.Interval
	}
	if config.Render != nil {
		g.Render = *config.Render
	}
	if config.TLDR != nil {
		g.TLDR = *config.TLDR
	}
	if config.Namespace != nil {
		g.CurrentNamespace = *config.Namespace
	}
	if config.Pager != nil {
		g.Pager = *config.Pager
	}
	if config.Web != nil {
		g.WebMode = *config.Web
	}
	if config.WebAddr != nil {
		g.WebAddr = *config.WebAddr
	}
	if config.WebDir != nil {
		g.WebDir = *config.WebDir
	}
	l.Debug("config loaded")
	return nil
}
