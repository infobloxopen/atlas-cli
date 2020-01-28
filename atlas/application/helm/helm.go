package helm

import (
	"strings"
)

type Helm struct {
	name string
	path string
}

func New(name string) *Helm {
	return &Helm{
		name: replaceName(name),
		path: "helm",
	}
}

func replaceName(name string) string {
	replaceList := []string{".", "_", ":"}
	helmName := name
	for _, s := range replaceList {
		helmName = strings.ReplaceAll(name, s, "-")
	}
	return helmName
}

func (h Helm) GetName() string {
	return h.name
}

func (h Helm) GetPath() string {
	return h.path
}

func (h Helm) GetDirs() []string {
	return []string{
		h.path,
		h.path + "/" + h.name,
		h.path + "/" + h.name + "/templates",
	}
}

func (h Helm) GetFiles(withDatabase bool) map[string]string {
	const (
		commonDir = iota
		templateDir
	)
	helmPath := h.GetDirs()[1:]
	tmplPath := []string{
		"templates/helm",
		"templates/helm/templates",
	}

	helmExt := ".yaml"
	tmplExt := ".yaml.gotmpl"

	listFiles := map[string]string{
		h.path + "/tpl.helm.properties":                    tmplPath[commonDir] + "/tpl.helm.properties.gotmpl",
		helmPath[commonDir] + "/.helmignore":               tmplPath[commonDir] + "/.helmignore.gotmpl",
		helmPath[commonDir] + "/Chart" + helmExt:           tmplPath[commonDir] + "/Chart" + tmplExt,
		helmPath[commonDir] + "/values" + helmExt:          tmplPath[commonDir] + "/values" + tmplExt,
		helmPath[commonDir] + "/minikube-values" + helmExt: tmplPath[commonDir] + "/minikube-values" + tmplExt,
		helmPath[templateDir] + "/_helpers.tpl":            tmplPath[templateDir] + "/_helpers.tpl.gotmpl",
		helmPath[templateDir] + "/deployment" + helmExt:    tmplPath[templateDir] + "/deployment" + tmplExt,
		helmPath[templateDir] + "/ingress" + helmExt:       tmplPath[templateDir] + "/ingress" + tmplExt,
		helmPath[templateDir] + "/NOTES.txt":               tmplPath[templateDir] + "/NOTES.txt.gotmpl",
		helmPath[templateDir] + "/service" + helmExt:       tmplPath[templateDir] + "/service" + tmplExt,
		helmPath[templateDir] + "/secrets" + helmExt:       tmplPath[templateDir] + "/secrets" + tmplExt,
	}

	if withDatabase {
		listFiles[helmPath[templateDir]+"/database"+helmExt] = tmplPath[templateDir] + "/database" + tmplExt
		listFiles[helmPath[templateDir]+"/migrations"+helmExt] = tmplPath[templateDir] + "/migrations" + tmplExt
	}

	return listFiles
}
