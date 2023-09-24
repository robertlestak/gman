package output

import (
	"git.shdw.tech/shdw.tech/gman/pkg/gman"
	"github.com/go-jose/go-jose/v3/json"
	"github.com/rodaine/table"
	"gopkg.in/yaml.v3"
)

func printAppsJSON(m *gman.Gman, apps []gman.App) error {
	jd, err := json.Marshal(apps)
	if err != nil {
		return err
	}
	println(string(jd))
	return nil
}

func printAppsYAML(m *gman.Gman, apps []gman.App) error {
	yd, err := yaml.Marshal(apps)
	if err != nil {
		return err
	}
	println(string(yd))
	return nil
}

func printAppsText(m *gman.Gman, apps []gman.App) error {
	if len(apps) == 0 {
		println("No apps found")
		return nil
	}
	if m.CurrentNamespace == "" {
		tbl := table.New("Namespace", "Name")
		for _, app := range apps {
			tbl.AddRow(app.Namespace, app.Name)
		}
		tbl.Print()
	} else {
		tbl := table.New("Name")
		for _, app := range apps {
			tbl.AddRow(app.Name)
		}
		tbl.Print()
	}
	return nil
}

func PrintApps(m *gman.Gman, apps []gman.App, output OutputType) error {
	switch output {
	case Text:
		return printAppsText(m, apps)
	case JSON:
		return printAppsJSON(m, apps)
	case YAML:
		return printAppsYAML(m, apps)
	}
	return nil
}

func stringInSlice(s string, ss []string) bool {
	for _, s2 := range ss {
		if s == s2 {
			return true
		}
	}
	return false
}

func printNamespacesJSON(m *gman.Gman, apps []gman.App) error {
	var namespaces []string
	for _, app := range apps {
		if !stringInSlice(app.Namespace, namespaces) {
			namespaces = append(namespaces, app.Namespace)
		}
	}
	jd, err := json.Marshal(namespaces)
	if err != nil {
		return err
	}
	println(string(jd))
	return nil
}

func printNamespacesYAML(m *gman.Gman, apps []gman.App) error {
	var namespaces []string
	for _, app := range apps {
		if !stringInSlice(app.Namespace, namespaces) {
			namespaces = append(namespaces, app.Namespace)
		}
	}
	yd, err := yaml.Marshal(namespaces)
	if err != nil {
		return err
	}
	println(string(yd))
	return nil
}

func printNamespacesText(m *gman.Gman, apps []gman.App) error {
	if len(apps) == 0 {
		println("No apps found")
		return nil
	}
	var namespaces []string
	for _, app := range apps {
		if !stringInSlice(app.Namespace, namespaces) {
			namespaces = append(namespaces, app.Namespace)
		}
	}
	tbl := table.New("Namespace")
	for _, namespace := range namespaces {
		tbl.AddRow(namespace)
	}
	tbl.Print()
	return nil
}

func PrintNamespaces(m *gman.Gman, apps []gman.App, output OutputType) error {
	switch output {
	case Text:
		return printNamespacesText(m, apps)
	case JSON:
		return printNamespacesJSON(m, apps)
	case YAML:
		return printNamespacesYAML(m, apps)
	}
	return nil
}
