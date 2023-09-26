package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"git.shdw.tech/shdw.tech/gman/internal/output"
	"git.shdw.tech/shdw.tech/gman/pkg/gman"
	"git.shdw.tech/shdw.tech/gman/pkg/release"
	log "github.com/sirupsen/logrus"
)

var (
	// Version is the version of the application
	Version        = "dev"
	gmancmd        = flag.NewFlagSet("gman", flag.ExitOnError)
	dir            = gmancmd.String("config", "~/.gman", "local directory")
	namespace      = gmancmd.String("n", "default", "namespace")
	allNamespaces  = gmancmd.Bool("A", false, "all namespaces")
	showNamespaces = gmancmd.Bool("ns", false, "list namespaces")
	tldr           = gmancmd.Bool("t", false, "show tldr")
	outputType     = gmancmd.String("o", "text", "output format for lists. text, json, yaml")
	render         = gmancmd.Bool("render", true, "render markdown")
	pager          = gmancmd.String("pager", "less", "pager")
	repo           = gmancmd.String("repo", "", "git repo")
	branch         = gmancmd.String("branch", "main", "git branch")
	updateInterval = gmancmd.String("interval", "24h", "update interval")
	forceUpdate    = gmancmd.Bool("pull", false, "update repo now")
	search         = gmancmd.String("s", "", "search")
	version        = gmancmd.Bool("version", false, "show version")
	printDir       = gmancmd.Bool("dir", false, "print man dir instead of showing contents")
	openURL        = gmancmd.Bool("open", false, "open url on get failure")
	releases       = gmancmd.Bool("r", false, "show releases")
	notifyReleases = gmancmd.Bool("notify", true, "notify on new releases")
	web            = gmancmd.Bool("web", false, "run web server")
	webAddr        = gmancmd.String("web-addr", ":8080", "web server address")
	webDir         = gmancmd.String("web-dir", "~/.gman/web", "web server directory.")
)

func init() {
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		ll = log.InfoLevel
	}
	log.SetLevel(ll)
}

func replaceTilde(path string) string {
	if path[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home + path[1:]
	}
	path = os.ExpandEnv(path)
	return path
}

func outputApp(m *gman.Gman, app *gman.App, printDir *bool) {
	// if printDir is set, print the app dir and exit
	if *printDir {
		fmt.Print(app.Dir)
		return
	}
	// if the user wants to see the tldr and one exists, show it and exit
	if m.TLDR && app.ShortFile != nil {
		tl, err := app.TLDR()
		if err != nil {
			if strings.HasPrefix(err.Error(), "get error") {
				// all we are going to return is the URL, so don't
				// try to render it
				m.Render = false
			} else {
				log.Fatal(err)
			}
		}
		if err := output.Print(m.Render, m.Pager, tl); err != nil {
			log.Fatal(err)
		}
		return
	}
	// if the user wants to see the readme and one exists, show it and exit
	if app.ReadmeFile != nil {
		rd, err := app.Readme()
		if err != nil {
			if strings.HasPrefix(err.Error(), "get error") {
				// all we are going to return is the URL, so don't
				// try to render it
				m.Render = false
			} else {
				log.Fatal(err)
			}
		}
		if err := output.Print(m.Render, m.Pager, rd); err != nil {
			log.Fatal(err)
		}
		return
	}
}

func releasesCmd(m *gman.Gman) {
	// load current releases
	if err := m.LoadReleases(); err != nil {
		log.Fatal(err)
	}
	// if we have an arg, show the release
	if len(gmancmd.Args()) == 1 {
		// get the release name from the args
		releaseName := gmancmd.Args()[0]
		// find the release
		release, err := m.GetRelease(releaseName)
		if err != nil {
			// if the release is not found, return an error
			log.Fatal(err)
		}
		// get the release readme
		rd, err := release.Readme()
		if err != nil {
			if strings.HasPrefix(err.Error(), "get error") {
				m.Render = false
			} else {
				log.Fatal(err)
			}
		}
		// if the release is found, show it
		if err := output.Print(m.Render, m.Pager, rd); err != nil {
			log.Fatal(err)
		}
		return
	}
	// if we want to search, do it and exit
	if *search != "" {
		rs := m.SearchReleases(*search)
		// if there is only one release, show it
		if len(rs) == 1 {
			rd, err := rs[0].Readme()
			if err != nil {
				if strings.HasPrefix(err.Error(), "get error") {
					m.Render = false
				} else {
					log.Fatal(err)
				}
			}
			if err := output.Print(m.Render, m.Pager, rd); err != nil {
				log.Fatal(err)
			}
			return
		}
		// otherwise, show all releases and let the user choose
		if err := output.PrintReleases(rs, output.OutputType(*outputType)); err != nil {
			log.Fatal(err)
		}
		return
	}
	// otherwise, list all releases
	rs := m.ListReleases()
	if rs != nil {
		if err := output.PrintReleases(rs, output.OutputType(*outputType)); err != nil {
			log.Fatal(err)
		}
	}
}

func searchCmd(m *gman.Gman) {
	apps := m.SearchApps(m.CurrentNamespace, *search)
	// if there is only one app, show it
	if len(apps) == 1 {
		outputApp(m, &apps[0], printDir)
		return
	}
	// otherwise, show all apps and let the user choose
	if err := output.PrintApps(m, apps, output.OutputType(*outputType)); err != nil {
		log.Fatal(err)
	}
}

func checkForUpdates(m *gman.Gman) {
	// load current releases
	if err := m.LoadReleases(); err != nil {
		log.Fatal(err)
	}
	currentReleases := m.ListReleases()
	// update the git repo
	if err := m.GitUpdate(); err != nil {
		log.Fatal(err)
	}
	// now, reload the releases
	if err := m.LoadReleases(); err != nil {
		log.Fatal(err)
	}
	// now, get the releases again
	rs := m.ListReleases()
	// check for new releases
	newReleases := release.NewReleases(currentReleases, rs)
	// if we have new releases, and the user wants to be notified, do it
	if newReleases != nil && len(newReleases) > 0 && m.NotifyOnNewRelease {
		for _, r := range newReleases {
			rd, err := r.Readme()
			if err != nil {
				if strings.HasPrefix(err.Error(), "get error") {
					m.Render = false
				} else {
					log.Fatal(err)
				}
			}
			if err := output.Print(m.Render, "", rd); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func webCmd(m *gman.Gman) {
	log.Info("starting web server. press ctrl-c to exit")
	if err := m.Server(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	logLevel := gmancmd.String("log", log.GetLevel().String(), "log level")
	gmancmd.Parse(os.Args[1:])
	ll, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(ll)
	l := log.WithFields(log.Fields{
		"version": Version,
		"fn":      "main",
	})
	l.Debug("starting")
	if *version {
		fmt.Printf("gman version %s", Version)
		return
	}
	gman.OpenURLOnGetFailure = *openURL
	release.OpenURLOnGetFailure = *openURL
	gitDir := replaceTilde(*dir)
	webDir := replaceTilde(*webDir)
	m := &gman.Gman{
		ConfigDir:          gitDir,
		CurrentNamespace:   *namespace,
		ForceUpdate:        *forceUpdate,
		NotifyOnNewRelease: *notifyReleases,
		Pager:              *pager,
		Render:             *render,
		TLDR:               *tldr,
		WebMode:            *web,
		WebAddr:            *webAddr,
		WebDir:             webDir,
	}
	dur, err := time.ParseDuration(*updateInterval)
	if err != nil {
		log.WithError(err).Fatal("invalid update interval")
	}
	m.UpdateInterval = dur
	m.Repo = &gman.Repo{
		URL:    *repo,
		Branch: *branch,
	}
	m.LoadConfig()
	m.LocalDir = path.Join(m.ConfigDir, m.RepoDir())
	// if we want to run the web server, do it and exit
	if m.WebMode {
		webCmd(m)
		return
	}
	if *allNamespaces {
		m.CurrentNamespace = ""
	}
	if m.Repo == nil || m.Repo.URL == "" {
		log.Fatal("no repo specified")
	}
	// check for updates and handle new releases
	checkForUpdates(m)
	// if we want to operate on releases, do it and exit
	if *releases {
		releasesCmd(m)
		return
	}
	// first, load all apps into memory
	if err := m.LoadApps(); err != nil {
		log.Fatal(err)
	}
	// if we only want to list the namespaces, do it and exit
	if *showNamespaces {
		l.Debug("listing namespaces")
		// set the current namespace to "" so we get all namespaces
		apps := m.ListApps("")
		if err := output.PrintNamespaces(m, apps, output.OutputType(*outputType)); err != nil {
			log.Fatal(err)
		}
		return
	}
	// if we want to search, do it and exit
	if *search != "" {
		l.Debug("searching apps")
		searchCmd(m)
		return
	}
	// if we only want to list apps, do it and exit
	if len(gmancmd.Args()) == 0 {
		l.Debug("listing apps")
		apps := m.ListApps(m.CurrentNamespace)
		if err := output.PrintApps(m, apps, output.OutputType(*outputType)); err != nil {
			log.Fatal(err)
		}
		return
	}
	// if there is only one arg, show the app
	if len(gmancmd.Args()) == 1 {
		// get the app name from the args
		appName := gmancmd.Args()[0]
		// find the app
		app, err := m.GetApp(m.CurrentNamespace, appName)
		if err != nil {
			// if the app is not found in the current namespace, try all namespaces
			if m.CurrentNamespace != "" {
				app, err = m.GetApp("", appName)
				if err != nil {
					// if the app is not found, return an error
					log.Fatal(err)
				}
			} else {
				// if the app is not found, return an error
				log.Fatal(err)
			}
		}
		// if the app is found, show it
		outputApp(m, app, printDir)
	}
}
