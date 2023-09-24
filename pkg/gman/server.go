package gman

import (
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"git.shdw.tech/shdw.tech/gman/internal/web"
	log "github.com/sirupsen/logrus"
)

func (g *Gman) initWeb() error {
	// if webdir doesnt exist, create it
	if _, err := os.Stat(g.WebDir); os.IsNotExist(err) {
		if err := web.WriteWebContent(g.WebDir); err != nil {
			return err
		}
	}
	cmd := exec.Command("npm", "install")
	cmd.Dir = g.WebDir
	if log.GetLevel() >= log.DebugLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (g *Gman) buildWeb() error {
	l := log.WithField("fn", "buildWeb")
	l.Debug("building web")
	// build the web
	cmd := exec.Command("npm", "run", "build")
	url, err := url.Parse(g.Repo.URL)
	if err != nil {
		return err
	}
	p := url.Path
	// remove any extension from the path
	p = p[:len(p)-len(filepath.Ext(p))]
	// split p on / and get the last element
	name := path.Base(p)
	cmd.Env = []string{
		"NODE_ENV=production",
		"SITE_TITLE=" + name,
		"RELEASES_DIR=" + path.Join(g.ConfigDir, g.RepoDir()) + "/releases",
		"DOCS_DIR=" + path.Join(g.ConfigDir, g.RepoDir()) + "/docs",
		"GIT_REPO=" + g.Repo.URL,
	}
	l.Debugf("env: %v", cmd.Env)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Dir = g.WebDir
	if log.GetLevel() >= log.DebugLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()
	if err != nil {
		return err
	}
	l.Debug("web built")
	return nil
}

func (g *Gman) serverUpdater() {
	l := log.WithField("fn", "serverUpdater")
	l.Debug("init web")
	if err := g.initWeb(); err != nil {
		l.Fatal(err)
	}
	l.Debug("web inited")
	for {
		l.Info("updating git")
		if err := g.GitUpdate(); err != nil {
			l.Fatal(err)
		}
		l.Debug("loading apps")
		if err := g.LoadApps(); err != nil {
			l.Error(err)
		}
		l.Debug("loading releases")
		if err := g.LoadReleases(); err != nil {
			l.Error(err)
		}
		l.Info("building web")
		if err := g.buildWeb(); err != nil {
			l.Fatal(err)
		}
		l.Info("web built")
		// sleep
		l.Debug("sleeping")
		time.Sleep(g.UpdateInterval)
	}
}

func (g *Gman) Server() error {
	if g.WebDir == "" {
		g.WebDir = path.Join(g.ConfigDir, "web")
	}
	// we are going to set the webdir contents
	// from an embedded filesystem, so clear out
	// whatever is there now, if anything
	os.RemoveAll(g.WebDir)
	go g.serverUpdater()
	staticDir := filepath.Join(g.WebDir, "build")
	http.Handle("/", http.FileServer(http.Dir(staticDir)))
	if err := http.ListenAndServe(g.WebAddr, nil); err != nil {
		return err
	}
	return nil
}
