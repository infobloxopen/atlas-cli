package update

import (
	"flag"
	"fmt"
	"github.com/infobloxopen/atlas-cli/atlas/application"
	"os"
)

type StructureError struct {
	Message string
}

func (e *StructureError) Error() string {
	return e.Message
}

const (
	// the full set of command names
	commandUpdateApp = "update-app"
)

var (
	// flag set for initializing the application
	updateFlagSet = flag.NewFlagSet(commandUpdateApp, flag.ExitOnError)
)

type Update struct{}

func (u Update) GetName() string {
	return commandUpdateApp
}

func (u Update) GetFlagset() *flag.FlagSet {
	return updateFlagSet
}

func (u Update) Run() error {
	if err := u.CheckFileStructure(); err != nil {
		return err
	}

	app := application.Application{}
	if err := app.Update(); err != nil {
		return err
	}

	return nil
}

func (u *Update) CheckFileStructure() error {
	wd, err := os.Getwd()
	if err != nil {
		return &StructureError{err.Error()}
	}

	if _, err := os.Stat(fmt.Sprintf("%s/Makefile.vars", wd)); os.IsNotExist(err) {
		return &StructureError{fmt.Sprintf("File %s/Makefile.vars not exist", wd)}
	}

	return nil
}
