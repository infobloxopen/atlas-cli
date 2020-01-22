package bootstrap

import (
	"errors"
	"flag"
	"fmt"
	"github.com/infobloxopen/atlas-cli/atlas/application"
	"github.com/infobloxopen/atlas-cli/atlas/application/helm"
	"github.com/infobloxopen/atlas-cli/atlas/utill"
	"go/build"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// the full set of command names
	commandInitApp = "init-app"

	// the full set of flag names
	flagAppName      = "name"
	flagRegistryName = "registry"
	flagWithGateway  = "gateway"
	flagWithDebug    = "debug"
	flagWithDatabase = "db"
	flagWithHealth   = "health"
	flagWithPubsub   = "pubsub"
	flagWithHelm     = "helm"
	flagExpandName   = "expand"
)

var (
	// flag set for initializing the application
	initialize         = flag.NewFlagSet(commandInitApp, flag.ExitOnError)
	initializeName     = initialize.String(flagAppName, "", "the application name (required)")
	initializeRegistry = initialize.String(flagRegistryName, "", "the Docker registry (optional)")
	initializeGateway  = initialize.Bool(flagWithGateway, false, "generate project with a gRPC gateway (default false)")
	initializeDebug    = initialize.Bool(flagWithDebug, false, "print debug statements during intialization (default false)")
	initializeDatabase = initialize.Bool(flagWithDatabase, false, "initialize the application with database folders")
	initializeHealth   = initialize.Bool(flagWithHealth, false, "initialize the application with internal health checks")
	initializePubsub   = initialize.Bool(flagWithPubsub, false, "initialize the application with a pubsub example")
	initializeHelm     = initialize.Bool(flagWithHelm, false, "initialize the application with the helm charts")
	initializeExpand   = initialize.String(flagExpandName, "", "the name of the input file for the `expand` command (optional)")
)

// bootstrap implements the command interface for project intialization
type Bootstrap struct{}

func (b Bootstrap) GetName() string { return "init-app" }

func (b Bootstrap) GetFlagset() *flag.FlagSet { return initialize }

func (b Bootstrap) Run() error {
	if *initializeName == "" {
		return initializationError{
			errors.New("application name is required"),
		}
	}
	wd, err := os.Getwd()
	if err != nil {
		return initializationError{err: err}
	}
	root, err := utill.ProjectRoot(build.Default.GOPATH, wd)
	if err != nil {
		return initializationError{err: err}
	}

	app := application.Application{
		Name:         *initializeName,
		Registry:     *initializeRegistry,
		Root:         root,
		WithGateway:  *initializeGateway,
		WithDatabase: *initializeDatabase,
		WithHealth:   *initializeHealth,
		WithPubsub:   *initializePubsub,
		WithHelm: *initializeHelm,
		ExpandName:   *initializeExpand,
	}

	if app.WithHelm {
		app.Helm = helm.New(app.Name)
	}

	if err := app.Initialize(); err != nil {
		return initializationError{err: err}
	}

	if app.ExpandName != "" {
		if err := expandResource(app.Name, app.ExpandName, app.WithDatabase); err != nil {
			return err
		}
		if err := CombineFiles("pkg/pb/service.proto", "pkg/pb/" + app.Name + ".proto"); err != nil {
				return err
		}
		if err := CombineFiles("pkg/svc/zserver.go", "pkg/svc/servers.go"); err != nil {
			return err
		}
	}

	if err := generateProtobuf(); err != nil {
		return err
	}
	if err := initDep(); err != nil {
		return err
	}
	if err := resolveImports(app.GetDirectories()); err != nil {
		return err
	}
	if err := initRepo(); err != nil {
		return err
	}

	return nil
}

type initializationError struct{ err error }

func (e initializationError) Error() string {
	return fmt.Sprintf("Unable to initialize application: %s", e.err.Error())
}

func runCommand(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	if *initializeDebug {
		cmd.Stderr = os.Stdout
		cmd.Stdout = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// generateProtobuf calls "make protobuf" to render initial .pb files
func generateProtobuf() error {
	fmt.Print("Generating protobuf files... ")
	if err := runCommand("make", "protobuf"); err != nil {
		return err
	}
	fmt.Println("done!")
	return nil
}

// initDep calls "dep init" to generate .toml files
func initDep() error {
	fmt.Print("Starting dep project... ")
	if err := runCommand("dep", "init"); err != nil {
		return err
	}
	fmt.Println("done!")
	return nil
}

// resolveImports resolves imports for a given set of a packages
func resolveImports(dirs []string) error {
	fmt.Print("Resolving imports... ")
	for _, dir := range dirs {
		if err := filepath.Walk(dir, resolveFileImports); err != nil {
			return err
		}
	}
	fmt.Println("done!")
	return nil
}

// resolveFileImports determines missing import paths for a given go file and
// also fixes any formatting issues
func resolveFileImports(path string, f os.FileInfo, err error) error {
	if err == nil && isGoFile(f) {
		withImports, err := imports.Process(path, nil, nil)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(path, withImports, 0); err != nil {
			return err
		}
	}
	return nil
}

func isGoFile(f os.FileInfo) bool {
	// ignore non-Go files
	name := f.Name()
	return !f.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

// initRepo initializes new applications as a git repository
func initRepo() error {
	fmt.Print("Initializing git repo... ")
	if err := runCommand("git", "init"); err != nil {
		return err
	}
	if err := runCommand("git", "add", "*"); err != nil {
		return err
	}
	if err := runCommand("git", "commit", "-m", "Initial commit"); err != nil {
		return err
	}
	fmt.Println("done!")
	return nil
}
