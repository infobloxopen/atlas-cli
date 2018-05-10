package commands

import (
	"flag"

	"github.com/infobloxopen/atlas-cli/atlas/commands/bootstrap"
)

// Command generically represents a command that is runnable via the atlas
// command-line tool (e.g. init-app)
type Command interface {
	GetName() string
	GetFlagset() *flag.FlagSet
	Run() error
}

// GetCommandSet returns a mapping between command names and commands
func GetCommandSet() map[string]Command {
	cmdBootstrap := bootstrap.Bootstrap{}
	return map[string]Command{
		cmdBootstrap.GetName(): cmdBootstrap,
	}
}

// GetCommandNames returns a list of all the command names
func GetCommandNames() []string {
	cmdBootstrap := bootstrap.Bootstrap{}
	return []string{cmdBootstrap.GetName()}
}
