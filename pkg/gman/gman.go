package gman

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"git.shdw.tech/shdw.tech/gman/internal/utils"
	"git.shdw.tech/shdw.tech/gman/pkg/release"
)

var (
	OpenURLOnGetFailure = false
)

type Gman struct {
	Repo               *Repo
	ConfigDir          string
	LocalDir           string
	Pager              string
	UpdateInterval     time.Duration
	CurrentNamespace   string
	ForceUpdate        bool
	NotifyOnNewRelease bool
	Render             bool
	TLDR               bool

	WebMode bool
	WebAddr string
	WebDir  string

	Apps     map[string][]App
	Releases []release.Release
}

type App struct {
	Namespace   string  `json:"namespace" yaml:"namespace"`
	Name        string  `json:"name" yaml:"name"`
	Dir         string  `json:"dir" yaml:"dir"`
	ReadmeFile  *string `json:"readmeFile" yaml:"readmeFile"`
	ShortFile   *string `json:"shortFile" yaml:"shortFile"`
	ExamplesDir *string `json:"examplesDir" yaml:"examplesDir"`
}

func (g *Gman) ListApps(namespace string) []App {
	if namespace == "" {
		var apps []App
		for _, a := range g.Apps {
			apps = append(apps, a...)
		}
		sort.Slice(apps, func(i, j int) bool {
			return apps[i].Name < apps[j].Name
		})
		return apps
	}
	// sort by namespace and name
	sort.Slice(g.Apps[namespace], func(i, j int) bool {
		return g.Apps[namespace][i].Name < g.Apps[namespace][j].Name
	})
	sort.Slice(g.Apps[namespace], func(i, j int) bool {
		return g.Apps[namespace][i].Namespace < g.Apps[namespace][j].Namespace
	})
	return g.Apps[namespace]
}

func (g *Gman) LoadReleases() error {
	if g.LocalDir == "" {
		return errors.New("local dir not set")
	}
	// ensure local dir exists
	_, err := os.Stat(g.LocalDir)
	if os.IsNotExist(err) {
		return nil
	}
	// load releases
	releaseDir := filepath.Join(g.LocalDir, "releases")
	// ensure local dir exists
	_, err = os.Stat(releaseDir)
	if os.IsNotExist(err) {
		return nil
	}
	rs, err := release.LoadReleases(releaseDir)
	if err != nil {
		return err
	}
	g.Releases = rs
	return nil
}

func (g *Gman) ListReleases() []release.Release {
	return g.Releases
}

func (g *Gman) GetRelease(releaseName string) (*release.Release, error) {
	if g.LocalDir == "" {
		return nil, errors.New("local dir not set")
	}
	releaseDir := filepath.Join(g.LocalDir, "releases")
	// ensure local dir exists
	_, err := os.Stat(releaseDir)
	if os.IsNotExist(err) {
		return nil, errors.New("local dir does not exist")
	}
	rs, err := release.LoadReleases(releaseDir)
	if err != nil {
		return nil, err
	}
	for _, release := range rs {
		if release.Name == releaseName {
			return &release, nil
		}
	}
	return nil, errors.New("release not found")
}

func appsSliceContains(apps []App, app App) bool {
	for _, a := range apps {
		if a.Name == app.Name {
			return true
		}
	}
	return false
}

func releaseSliceContains(releases []release.Release, release release.Release) bool {
	for _, r := range releases {
		if r.Name == release.Name {
			return true
		}
	}
	return false
}

func (g *Gman) SearchApps(namespace string, search string) []App {
	l := log.WithField("fn", "SearchApps")
	l.Debug("searching apps")
	// Don't open URLs on get failure when searching
	OpenURLOnGetFailure = false
	var apps []App
	for _, app := range g.ListApps(namespace) {
		l.WithField("app", app.Name).Debug("checking app")
		if strings.Contains(app.Name, search) {
			l.Debug("app found")
			apps = append(apps, app)
		}
		// check inside the readme
		if app.ReadmeFile != nil {
			rd, err := app.Readme()
			if err != nil {
				l.WithError(err).Debug("error getting readme")
				continue
			}
			if utils.StringSearch(rd, search) && !appsSliceContains(apps, app) {
				l.Debug("app found")
				apps = append(apps, app)
			}
		}
		// check inside the tldr
		if app.ShortFile != nil {
			tl, err := app.TLDR()
			if err != nil {
				l.WithError(err).Debug("error getting tldr")
				continue
			}
			if utils.StringSearch(tl, search) && !appsSliceContains(apps, app) {
				l.Debug("app found")
				apps = append(apps, app)
			}
		}
	}
	// if we found no apps, try across all namespaces
	if len(apps) == 0 {
		l.Debug("no apps found, searching all namespaces")
		apps = g.SearchApps("", search)
	}
	l.Debug("apps searched")
	return apps
}

func (g *Gman) SearchReleases(search string) []release.Release {
	// Don't open URLs on get failure when searching
	release.OpenURLOnGetFailure = false
	var rr []release.Release
	rs := g.ListReleases()
	if rs == nil {
		return nil
	}
	for _, r := range rs {
		if strings.Contains(r.Name, search) {
			rr = append(rr, r)
		}
		// check inside the readme
		if r.ReadmeFile != nil {
			rd, err := r.Readme()
			if err != nil {
				continue
			}
			if utils.StringSearch(rd, search) && !releaseSliceContains(rr, r) {
				rr = append(rr, r)
			}
		}
	}
	return rr
}

func (g *Gman) LoadApps() error {
	l := log.WithField("fn", "LoadApps")
	l.Debug("loading apps")
	if g.LocalDir == "" {
		l.Error("local dir not set")
		return errors.New("local dir not set")
	}
	// ensure local dir exists
	_, err := os.Stat(g.LocalDir)
	if os.IsNotExist(err) {
		l.Error("local dir does not exist")
		return errors.New("local dir does not exist")
	}
	if g.Apps == nil {
		g.Apps = make(map[string][]App)
	}
	// walk the local dir
	root := filepath.Join(g.LocalDir, "docs")
	l.WithField("root", root).Debug("walking local dir")
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			l.WithError(err).Error("error walking local dir")
			return err
		}

		// check for the README.md inside the app directory
		if strings.EqualFold(info.Name(), "README.md") {
			rel, _ := filepath.Rel(root, path)
			parts := strings.Split(rel, string(os.PathSeparator))
			if len(parts) >= 2 {
				namespace := parts[0]
				appName := parts[1]
				readmeFile := path
				var shortFile *string
				var examplesDir *string
				// Check for ShortFile and ExamplesDir in the same app directory
				shortFilePath := filepath.Join(filepath.Dir(path), "TLDR.md")
				if _, err := os.Stat(shortFilePath); err == nil {
					shortFile = &shortFilePath
				}
				examplesDirPath := filepath.Join(filepath.Dir(path), "examples")
				if _, err := os.Stat(examplesDirPath); err == nil {
					examplesDir = &examplesDirPath
				}
				app := App{
					Namespace:   namespace,
					Name:        appName,
					Dir:         filepath.Dir(path),
					ReadmeFile:  &readmeFile,
					ShortFile:   shortFile,
					ExamplesDir: examplesDir,
				}
				g.Apps[namespace] = append(g.Apps[namespace], app)
			}
		}
		return nil
	})
	return err
}

func (g *Gman) GetApp(namespace, name string) (*App, error) {
	l := log.WithField("fn", "GetApp")
	l.Debug("getting app")
	if namespace == "" {
		l.Debug("namespace not set, searching all namespaces")
		// first, check if the app exists in the default namespace
		if g.Apps["default"] != nil {
			for _, app := range g.Apps["default"] {
				if app.Name == name {
					l.Debug("app found")
					return &app, nil
				}
			}
		}
		l.Debug("app not found in default namespace")
		// next, check if the app exists in any namespace
		for _, apps := range g.Apps {
			for _, app := range apps {
				if app.Name == name {
					l.Debug("app found")
					return &app, nil
				}
			}
		}
		l.Debug("app not found in any namespace")
		// app does not exist in any namespace
		return nil, errors.New("app not found")
	}
	l.Debug("namespace set, searching only in namespace")
	// explicitly check the namespace for the app
	for _, app := range g.Apps[namespace] {
		if app.Name == name {
			l.Debug("app found")
			return &app, nil
		}
	}
	l.Debug("app not found in namespace")
	return nil, errors.New("app not found")
}

func (a *App) Readme() (string, error) {
	l := log.WithField("fn", "Readme")
	l.Debug("getting readme")
	if a.ReadmeFile == nil {
		l.Error("readme file not set")
		return "", errors.New("readme file not set")
	}
	// read the file
	b, err := os.ReadFile(*a.ReadmeFile)
	if err != nil {
		l.WithError(err).Error("error reading readme file")
		return "", err
	}
	l.Debug("readme file read")
	if utils.IsOnlyUrl(string(b)) {
		l.Debug("readme file is only a url")
		res, err := utils.GetRemote(string(b))
		if err != nil {
			if OpenURLOnGetFailure {
				l.Debug("opening url")
				utils.OpenURL(string(b))
			}
			return string(b), err
		}
		b = []byte(res)
	}
	l.Debug("readme gotten")
	return string(b), nil
}

func (a *App) TLDR() (string, error) {
	if a.ShortFile == nil {
		return "", errors.New("tldr file not set")
	}
	// read the file
	b, err := os.ReadFile(*a.ShortFile)
	if err != nil {
		return "", err
	}
	if utils.IsOnlyUrl(string(b)) {
		res, err := utils.GetRemote(string(b))
		if err != nil {
			if OpenURLOnGetFailure {
				utils.OpenURL(string(b))
			}
			return string(b), err
		}
		b = []byte(res)
	}
	return string(b), nil
}
