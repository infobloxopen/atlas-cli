package application

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/infobloxopen/atlas-cli/atlas/application/helm"
	"github.com/infobloxopen/atlas-cli/atlas/templates"
	"github.com/infobloxopen/atlas-cli/atlas/utill"
)

// Application models the data that the templates need to render files
type Application struct {
	Name               string
	Registry           string
	Root               string
	WithGateway        bool
	WithDatabase       bool
	WithHealth         bool
	WithGithub         bool
	WithMetrics        bool
	WithPubsub         bool
	WithHelm           bool
	WithProfiler       bool
	Helm               *helm.Helm
	ExpandName         string
	WithKind           bool
	WithDelve          bool
	WithSubscribeTopic string
	WithPublishTopic   string
}

// Initialize generates brand-new application
func (app Application) Initialize() error {
	if _, err := os.Stat(app.Name); !os.IsNotExist(err) {
		msg := fmt.Sprintf("directory '%v' already exists.", app.Name)
		return errors.New(msg)
	}
	if err := os.Mkdir(app.Name, os.ModePerm); err != nil {
		return err
	}
	if err := os.Chdir(app.Name); err != nil {
		return err
	}
	if err := app.initializeDirectories(); err != nil {
		return err
	}

	if err := app.initializeFiles(); err != nil {
		return err
	}

	return nil
}

// Update update application and bring latest features
func (app Application) Update() error {
	fmt.Print("Regenerating Makefile.common... ")
	if err := app.generateMakefileCommon(); err != nil {
		return err
	}

	fmt.Println("done")

	return nil
}

// initializeFiles generates each file for a new application
func (app Application) initializeFiles() error {
	fileInitializers := []func(Application) error{
		app.generateFiles(
			"templates/docker/Dockerfile.gotmpl",
			"templates/deploy/config.yaml.gotmpl",
			"templates/README.md.gotmpl",
			"templates/.gitignore.gotmpl",
			"templates/Makefile.gotmpl",
			"templates/Makefile.vars.gotmpl",
			"templates/Jenkinsfile.gotmpl",
			"templates/pkg/pb/service.proto.gotmpl",
			"templates/cmd/server/main.go.gotmpl",
			"templates/cmd/server/grpc.go.gotmpl",
			"templates/cmd/server/config.go.gotmpl",
			"templates/pkg/svc/zserver.go.gotmpl",
			"templates/pkg/svc/zserver_test.go.gotmpl",
		),
		Application.generateGoMod,
		Application.generateMakefileCommon,
	}
	if app.WithSubscribeTopic != "" || app.WithPublishTopic != "" {
		fileInitializers = append(fileInitializers, Application.generatePubsub, Application.generatePubsubTest)
	}
	if app.WithKind {
		fileInitializers = append(fileInitializers, Application.generateMakefileKind,
			Application.generateKindConfig, Application.generateKindConfigYaml,
			Application.generateKindConfigV119, Application.generateRedisNoPassword)
	}
	if app.WithDelve {
		fileInitializers = append(fileInitializers, Application.generateMakefileDebugger)
		fileInitializers = append(fileInitializers, Application.generateDockerfileDebug)
	}
	if app.WithGateway {
		fileInitializers = append(fileInitializers, Application.generateServerSwagger)
	}
	if app.WithDatabase {
		fileInitializers = append(fileInitializers, Application.generateMigrationFile)
	}
	if app.WithGithub {
		fileInitializers = append(fileInitializers, app.generateFiles(filterAssets("templates/.github")...))
	}
	if app.Helm != nil {
		fileInitializers = append(fileInitializers, Application.generateHelmCharts)
	}
	if app.WithProfiler {
		fileInitializers = append(fileInitializers, Application.generateServerProfiler)
	}

	for _, initializer := range fileInitializers {
		if err := initializer(app); err != nil {
			return err
		}
	}
	return nil
}

// initializeDirectories generates the directory tree for a new application
func (app Application) initializeDirectories() error {
	dirs := app.GetDirectories()
	for _, dir := range dirs {
		if err := os.MkdirAll(fmt.Sprintf("./%s", dir), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// GetDirectories returns a list of all project folders
func (app Application) GetDirectories() []string {
	dirnames := []string{
		"cmd",
		"cmd/server",
		"pkg",
		"pkg/pb",
		"pkg/svc",
		"docker",
		"deploy",
	}
	if app.WithKind {
		dirnames = append(dirnames,
			"kind",
		)
	}
	if app.WithSubscribeTopic != "" || app.WithPublishTopic != "" {
		dirnames = append(dirnames,
			"pkg/dapr",
		)
	}
	if app.WithDatabase {
		dirnames = append(dirnames,
			"db/migrations",
			"db/fixtures",
		)
	}
	if app.WithGithub {
		dirnames = append(dirnames,
			mapString(filterAssetDirs("templates/.github"), func(s string) string {
				return strings.TrimPrefix(s, "templates/")
			})...,
		)
	}
	if app.WithHelm {
		dirnames = append(dirnames,
			app.Helm.GetDirs()...,
		)
	}
	return dirnames
}

// generateFile creates a file by rendering a template
func (app Application) generateFile(filename, templatePath string) error {
	t := template.New("file").Funcs(template.FuncMap{
		"Title":    strings.Title,
		"Service":  utill.ServiceName,
		"URL":      utill.ServerURL,
		"Database": utill.DatabaseName,
		"Package":  utill.PackageName,
	})
	bytes, err := templates.Asset(templatePath)
	if err != nil {
		return err
	}
	t, err = t.Parse(string(bytes))
	if err != nil {
		return err
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := t.Execute(file, app); err != nil {
		return err
	}
	return err
}

func (app Application) generateKindConfig() error {
	return app.generateFile("kind/kind-config", "templates/kind/kind-config.gotmpl")
}

func (app Application) generateKindConfigYaml() error {
	return app.generateFile("kind/kind-config.yaml.in", "templates/kind/kind-config.yaml.in.gotmpl")
}

func (app Application) generateKindConfigV119() error {
	return app.generateFile("kind/kind-config-v1.19.0.yaml", "templates/kind/kind-config-v1.19.0.yaml.gotmpl")
}

func (app Application) generateRedisNoPassword() error {
	return app.generateFile("kind/redis-no-password.yaml.template", "templates/kind/redis-no-password.yaml.template.gotmpl")
}

// generateFiles is a helper for calling app.generateFile().
// It calls app.generateFile() for each given filename.
func (app Application) generateFiles(tmplPaths ...string) func(Application) error {
	return func(Application) error {
		for _, tmplPath := range tmplPaths {
			if err := app.generateFile(strings.TrimSuffix(strings.TrimPrefix(tmplPath, "templates/"), ".gotmpl"), tmplPath); err != nil {
				return err
			}
		}
		return nil
	}
}

func (app Application) generateDockerfileDebug() error {
	return app.generateFile("docker/Dockerfile.debug", "templates/docker/Dockerfile.debug.gotmpl")
}

func (app Application) generateMigrationFile() error {
	return app.generateFile("deploy/migrations.yaml", "templates/deploy/migrations.yaml.gotmpl")
}

func (app Application) generateGoMod() error {
	return app.generateFile("go.mod", "templates/go.mod.gotmpl")
}

func (app Application) generateMakefileKind() error {
	return app.generateFile("Makefile.kind", "templates/Makefile.kind.gotmpl")
}

func (app Application) generateMakefileDebugger() error {
	return app.generateFile("Makefile.remotedebug", "templates/Makefile.remotedebug.gotmpl")
}

func (app Application) generateMakefileCommon() error {
	return app.generateFile("Makefile.common", "templates/Makefile.common.gotmpl")
}

func (app Application) generateServerProfiler() error {
	return app.generateFile("cmd/server/profiler.go", "templates/cmd/server/profiler.go.gotmpl")
}

func (app Application) generateServerSwagger() error {
	return app.generateFile("cmd/server/swagger.go", "templates/cmd/server/swagger.go.gotmpl")
}

func (app Application) generatePubsub() error {
	return app.generateFile("pkg/dapr/pubsub.go", "templates/pkg/dapr/pubsub.go.gotmpl")
}

func (app Application) generatePubsubTest() error {
	return app.generateFile("pkg/dapr/pubsub_test.go", "templates/pkg/dapr/pubsub_test.go.gotmpl")
}

func (app Application) generateHelmCharts() error {
	for helmPath, tmplPath := range app.Helm.GetFiles(app.WithDatabase) {
		if err := app.generateFile(helmPath, tmplPath); err != nil {
			return err
		}
	}
	return nil
}

func filterString(vs []string, f func(string) bool) []string {
	vsf := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func mapString(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

// filterAssets returns a list of assets under the given prefix/path
func filterAssets(prefix string) []string {
	return filterString(templates.AssetNames(), func(s string) bool {
		return strings.HasPrefix(s, prefix)
	})
}

// filterAssets returns a list of directories that contain assets under the given prefix/path
func filterAssetDirs(prefix string) []string {
	assets := filterAssets(prefix)
	dirs := make(map[string]struct{}, len(assets))
	for _, asset := range assets {
		dirs[path.Dir(asset)] = struct{}{}
	}
	dedupedDirs := make([]string, len(dirs))
	for dir := range dirs {
		dedupedDirs = append(dedupedDirs, dir)
	}
	return dedupedDirs
}
