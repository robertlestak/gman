package output

import (
	"git.shdw.tech/shdw.tech/gman/pkg/release"
	"github.com/go-jose/go-jose/v3/json"
	"github.com/rodaine/table"
	"gopkg.in/yaml.v3"
)

func printReleasesJSON(releases []release.Release) error {
	jd, err := json.Marshal(releases)
	if err != nil {
		return err
	}
	println(string(jd))
	return nil
}

func printReleasesYAML(releases []release.Release) error {
	yd, err := yaml.Marshal(releases)
	if err != nil {
		return err
	}
	println(string(yd))
	return nil
}

func printReleasesText(releases []release.Release) error {
	if len(releases) == 0 {
		println("No releases found")
		return nil
	}
	tbl := table.New("Name", "Date")
	for _, release := range releases {
		tbl.AddRow(release.Name, release.Date.Format("2006-01-02 15:04:05"))
	}
	tbl.Print()
	return nil
}

func PrintReleases(releases []release.Release, output OutputType) error {
	switch output {
	case Text:
		return printReleasesText(releases)
	case JSON:
		return printReleasesJSON(releases)
	case YAML:
		return printReleasesYAML(releases)
	}
	return nil
}
