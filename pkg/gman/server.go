package gman

import (
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"git.shdw.tech/shdw.tech/gman/internal/utils"
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
	var editUrl string
	if strings.HasSuffix(g.Repo.URL, ".git") {
		editUrl = strings.TrimSuffix(g.Repo.URL, ".git")
	} else {
		editUrl = g.Repo.URL
	}
	editUrl = editUrl + "/blob/" + g.Repo.Branch
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, []string{
		"NODE_ENV=production",
		"SITE_TITLE=" + name,
		"RELEASES_DIR=" + path.Join(g.ConfigDir, g.RepoDir()) + "/releases",
		"DOCS_DIR=" + path.Join(g.ConfigDir, "web", "docs"),
		"GIT_REPO=" + g.Repo.URL,
		"GIT_REPO_EDIT_URL=" + editUrl,
	}...)
	l.Debugf("env: %v", cmd.Env)
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

func (g *Gman) RenderDocsToDisk() error {
	l := log.WithField("fn", "RenderDocsToDisk")
	l.Debug("rendering docs to disk")
	// first, create a directory to hold our rendered docs
	renderedDocsDir := path.Join(g.ConfigDir, "web", "docs")
	// first, copy over everything as-is
	if err := utils.Copydir(renderedDocsDir, path.Join(g.LocalDir, "docs")); err != nil {
		return err
	}
	// next, render the docs
	for _, apps := range g.Apps {
		for _, app := range apps {
			if app.ReadmeFile == nil {
				continue
			}
			l.Debugf("rendering docs for %s/%s", app.Namespace, app.Name)
			// readme file
			if app.ReadmeFile != nil {
				readmeFile := *app.ReadmeFile
				l.Debugf("readme file: %s", readmeFile)
				newReameFile := strings.Replace(readmeFile, path.Join(g.LocalDir, "docs"), renderedDocsDir, 1)
				l.Debugf("new readme file: %s", newReameFile)
				// ensure parent dir exists
				if err := os.MkdirAll(filepath.Dir(newReameFile), 0755); err != nil {
					return err
				}
				// render the readme
				rd, err := app.Readme()
				if err != nil {
					// just copy the original file as-is
					if _, err := utils.CopyFile(readmeFile, newReameFile); err != nil {
						continue
					}
				} else {
					if err := os.WriteFile(newReameFile, []byte(rd), 0644); err != nil {
						return err
					}
				}
				l.Debugf("rendered readme file: %s", newReameFile)
			}
			if app.ShortFile != nil {
				// tldr file
				shortfile := *app.ShortFile
				l.Debugf("tldr file: %s", shortfile)
				newShortFile := strings.Replace(shortfile, path.Join(g.LocalDir, "docs"), renderedDocsDir, 1)
				l.Debugf("new tldr file: %s", newShortFile)
				// ensure parent dir exists
				if err := os.MkdirAll(filepath.Dir(newShortFile), 0755); err != nil {
					return err
				}
				// render the tldr
				tl, err := app.TLDR()
				if err != nil {
					// just copy the original file as-is
					if _, err := utils.CopyFile(shortfile, newShortFile); err != nil {
						continue
					}
				} else {

					if err := os.WriteFile(newShortFile, []byte(tl), 0644); err != nil {
						return err
					}
				}
				l.Debugf("rendered tldr file: %s", newShortFile)
			}
			// if there is an examples dir, just copy it over as-is
			if app.ExamplesDir != nil {
				examplesDir := *app.ExamplesDir
				l.Debugf("examples dir: %s", examplesDir)
				newExamplesDir := strings.Replace(examplesDir, path.Join(g.LocalDir, "docs"), renderedDocsDir, 1)
				l.Debugf("new examples dir: %s", newExamplesDir)
				// ensure parent dir exists
				if err := os.MkdirAll(newExamplesDir, 0755); err != nil {
					return err
				}
				// copy the examples dir
				if err := utils.Copydir(examplesDir, newExamplesDir); err != nil {
					return err
				}
				l.Debugf("copied examples dir: %s", newExamplesDir)
			}
		}
	}
	return nil
}

func (g *Gman) serverUpdater() {
	l := log.WithField("fn", "serverUpdater")
	l.Debug("init web")
	log.Info("initializing node environment...")
	if err := g.initWeb(); err != nil {
		l.Fatal(err)
	}
	l.Debug("web inited")
	for {
		l.Debug("updating git")
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
		if err := g.RenderDocsToDisk(); err != nil {
			l.Error(err)
		}
		l.Info("building web app...")
		if err := g.buildWeb(); err != nil {
			l.Fatal(err)
		}
		l.Info("web app built, ready to serve")
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
	// don't open the browser on get failure
	OpenURLOnGetFailure = false
	// set ServerMode to true
	ServerMode = true
	go g.serverUpdater()
	staticDir := filepath.Join(g.WebDir, "build")
	http.Handle("/", http.FileServer(http.Dir(staticDir)))
	log.Infof("server listening on %s", g.WebAddr)
	if err := http.ListenAndServe(g.WebAddr, nil); err != nil {
		return err
	}
	return nil
}
