package gman

import (
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

func (g *Gman) RepoDir() string {
	l := log.WithField("fn", "RepoDir")
	l.Debug("getting repo dir")
	// parse url
	l.Debugf("parsing url %s", g.Repo.URL)
	url, err := url.Parse(g.Repo.URL)
	if err != nil {
		return ""
	}
	domain := url.Hostname()
	path := url.Path
	// remove any extension from the path
	path = path[:len(path)-len(filepath.Ext(path))]
	p := filepath.Join("src", domain, path)
	l.Debugf("repo dir is %s", p)
	return p
}

func (g *Gman) GitClone() error {
	cmd := exec.Command("git", "clone", "-b", g.Repo.Branch, g.Repo.URL, g.LocalDir)
	if log.GetLevel() >= log.DebugLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		return err
	}
	if err := g.GitUpdateSubmodules(); err != nil {
		return err
	}
	return nil
}

func (g *Gman) GitPull() error {
	l := log.WithField("fn", "GitPull")
	l.Debug("pulling git repo")
	// update our local git repo to the latest and ensure we're on the right branch
	cmd := exec.Command("git", "pull", "origin", g.Repo.Branch)
	cmd.Dir = g.LocalDir
	l.WithField("dir", g.LocalDir).Debug("running git pull")
	if log.GetLevel() >= log.DebugLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Run()
	if err != nil {
		l.WithError(err).Error("error running git pull")
		return err
	}
	cmd = exec.Command("git", "checkout", g.Repo.Branch)
	cmd.Dir = g.LocalDir
	if log.GetLevel() >= log.DebugLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()
	if err != nil {
		l.WithError(err).Error("error running git checkout")
		return err
	}
	// git reset --hard origin/{branch}
	cmd = exec.Command("git", "reset", "--hard", "origin/"+g.Repo.Branch)
	cmd.Dir = g.LocalDir
	if log.GetLevel() >= log.DebugLevel {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err = cmd.Run()
	if err != nil {
		l.WithError(err).Error("error running git reset")
		return err
	}
	if err := g.GitUpdateSubmodules(); err != nil {
		l.WithError(err).Error("error running git update submodules")
		return err
	}
	return nil
}

func (g *Gman) LastUpdated() (time.Time, error) {
	// last pull is
	// stat -c %Y .git/FETCH_HEAD
	fetchHeadFile := filepath.Join(g.LocalDir, ".git", "FETCH_HEAD")
	stat, err := os.Stat(fetchHeadFile)
	if err != nil {
		return time.Time{}, err
	}
	return stat.ModTime(), nil
}

func (g *Gman) GitUpdateSubmodules() error {
	// update submodules
	cmd := exec.Command("git", "submodule", "update", "--init", "--recursive")
	cmd.Dir = g.LocalDir
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

func (g *Gman) GitUpdate() error {
	l := log.WithField("fn", "GitUpdate")
	l.Debug("updating git repo")
	// if repo doesn't exist, clone it
	if _, err := os.Stat(g.LocalDir); os.IsNotExist(err) {
		l.Debug("repo does not exist, cloning")
		return g.GitClone()
	}
	if g.ForceUpdate {
		l.Debug("force update set, pulling")
		return g.GitPull()
	}
	// if repo exists and we have updated within the update interval, do nothing
	// otherwise, pull
	lastUpdated, err := g.LastUpdated()
	if err != nil {
		l.Debug("unable to get last updated time, pulling")
		// if we can't get the last updated time, pull
		return g.GitPull()
	}
	if lastUpdated.IsZero() || g.UpdateInterval > 0 && lastUpdated.Add(g.UpdateInterval).Before(time.Now()) {
		l.Debug("last updated time is before update interval, pulling")
		return g.GitPull()
	}
	l.Debug("last updated time is after update interval, not pulling")
	return nil
}
