package application

import (
	"errors"
	"fmt"
	"github.com/infobloxopen/atlas-cli/atlas/templates"
	"github.com/infobloxopen/atlas-cli/atlas/utill"
	"os"
	"strings"
	"text/template"
)

// Application models the data that the templates need to render files
type Application struct {
	Name         string
	Registry     string
	Root         string
	WithGateway  bool
	WithDatabase bool
	WithHealth   bool
	WithPubsub   bool
	ExpandName   string
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
		Application.generateDockerfile,
		Application.generateDeployFile,
		Application.generateReadme,
		Application.generateGitignore,
		Application.generateMakefileVars,
		Application.generateMakefileCommon,
		Application.generateMakefile,
		Application.generateJenkinsfile,
		Application.generateProto,
		Application.generateServerMain,
		Application.generateServerGrpc,
		Application.generateConfig,
		Application.generateService,
		Application.generateServiceTest,
	}
	if app.WithGateway {
		fileInitializers = append(fileInitializers, Application.generateServerSwagger)
	}
	if app.WithDatabase {
		fileInitializers = append(fileInitializers, Application.generateMigrationFile)
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
	if app.WithDatabase {
		dirnames = append(dirnames,
			"db/migrations",
			"db/fixtures",
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

func (app Application) generateDockerfile() error {
	return app.generateFile("docker/Dockerfile", "templates/docker/Dockerfile.gotmpl")
}

func (app Application) generateDeployFile() error {
	return app.generateFile("deploy/config.yaml", "templates/deploy/config.yaml.gotmpl")
}

func (app Application) generateMigrationFile() error {
	return app.generateFile("deploy/migrations.yaml", "templates/deploy/migrations.yaml.gotmpl")
}

func (app Application) generateReadme() error {
	return app.generateFile("README.md", "templates/README.md.gotmpl")
}

func (app Application) generateGitignore() error {
	return app.generateFile(".gitignore", "templates/.gitignore.gotmpl")
}

func (app Application) generateMakefile() error {
	return app.generateFile("Makefile", "templates/Makefile.gotmpl")
}

func (app Application) generateMakefileVars() error {
	return app.generateFile("Makefile.vars", "templates/Makefile.vars.gotmpl")
}

func (app Application) generateMakefileCommon() error {
	return app.generateFile("Makefile.common", "templates/Makefile.common.gotmpl")
}

func (app Application) generateJenkinsfile() error {
	return app.generateFile("Jenkinsfile", "templates/Jenkinsfile.gotmpl")
}

func (app Application) generateProto() error {
	return app.generateFile("pkg/pb/service.proto", "templates/pkg/pb/service.proto.gotmpl")
}

func (app Application) generateServerMain() error {
	return app.generateFile("cmd/server/main.go", "templates/cmd/server/main.go.gotmpl")
}

func (app Application) generateServerGrpc() error {
	return app.generateFile("cmd/server/grpc.go", "templates/cmd/server/grpc.go.gotmpl")
}

func (app Application) generateServerSwagger() error {
	return app.generateFile("cmd/server/swagger.go", "templates/cmd/server/swagger.go.gotmpl")
}

func (app Application) generateConfig() error {
	return app.generateFile("cmd/server/config.go", "templates/cmd/server/config.go.gotmpl")
}

func (app Application) generateService() error {
	return app.generateFile("pkg/svc/zserver.go", "templates/pkg/svc/zserver.go.gotmpl")
}

func (app Application) generateServiceTest() error {
	return app.generateFile("pkg/svc/zserver_test.go", "templates/pkg/svc/zserver_test.go.gotmpl")
}
