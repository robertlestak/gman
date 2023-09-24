package release

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/mod/semver"

	"git.shdw.tech/shdw.tech/gman/internal/utils"
)

var (
	OpenURLOnGetFailure = false
)

type Release struct {
	Name       string
	Date       time.Time
	Dir        string  `json:"dir" yaml:"dir"`
	ReadmeFile *string `json:"readmeFile" yaml:"readmeFile"`
}

func SortBySemver(rs []Release) {
	// use semver.Sort to reorder the releases
	sort.Slice(rs, func(i, j int) bool {
		return semver.Compare(rs[i].Name, rs[j].Name) == -1
	})
	// reverse the order, to get the latest release first
	for i := len(rs)/2 - 1; i >= 0; i-- {
		opp := len(rs) - 1 - i
		rs[i], rs[opp] = rs[opp], rs[i]
	}
}

func LoadReleases(dir string) ([]Release, error) {
	if dir == "" {
		return nil, errors.New("local dir not set")
	}
	// ensure local dir exists
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, errors.New("local dir does not exist")
	}
	var rs []Release

	// walk the local dir
	root := filepath.Join(dir)
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// check for the README.md inside the app directory
		if strings.EqualFold(info.Name(), "README.md") {
			rel, _ := filepath.Rel(root, path)
			parts := strings.Split(rel, string(os.PathSeparator))
			if len(parts) >= 1 {
				releaseName := parts[0]
				readmeFile := path
				// check when the file was last modified
				lastModified := info.ModTime()
				// if the file was modified in the future, use the current time
				if lastModified.After(time.Now()) {
					lastModified = time.Now()
				}

				release := Release{
					Name:       releaseName,
					Dir:        filepath.Dir(path),
					ReadmeFile: &readmeFile,
					Date:       lastModified,
				}
				rs = append(rs, release)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// if the first release is a valid semver, assume all releases are and sort them
	if len(rs) > 0 && semver.IsValid(rs[0].Name) {
		SortBySemver(rs)
	}
	return rs, err
}

func (r *Release) Readme() (string, error) {
	if r.ReadmeFile == nil {
		return "", errors.New("readme file not set")
	}
	// read the file
	b, err := os.ReadFile(*r.ReadmeFile)
	if err != nil {
		return "", err
	}
	if utils.IsOnlyUrl(string(b)) {
		res, err := http.Get(string(b))
		if err != nil {
			if OpenURLOnGetFailure {
				utils.OpenURL(string(b))
			}
			return string(b), nil
		}
		defer res.Body.Close()
		bd, err := io.ReadAll(res.Body)
		if err != nil {
			return string(b), nil
		}
		b = bd
	}
	return string(b), nil
}

func NewReleases(current []Release, latest []Release) []Release {
	var newReleases []Release
	for _, l := range latest {
		found := false
		for _, c := range current {
			if l.Name == c.Name {
				found = true
				break
			}
		}
		if !found {
			newReleases = append(newReleases, l)
		}
	}
	// if the first release is a valid semver, assume all releases are and sort them
	if len(newReleases) > 0 && semver.IsValid(newReleases[0].Name) {
		SortBySemver(newReleases)
	}
	return newReleases
}
