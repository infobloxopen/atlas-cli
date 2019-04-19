# Atlas CLI

[![Build Status](https://travis-ci.org/infobloxopen/atlas-cli.svg?branch=master)](https://travis-ci.org/infobloxopen/atlas-cli) [![Go Report Card](https://goreportcard.com/badge/github.com/infobloxopen/atlas-cli)](https://goreportcard.com/report/github.com/infobloxopen/atlas-cli)

This command-line tool helps developers become productive on Atlas. It aims to provide a better development experience by reducing the initial time and effort it takes to build applications.

## Getting Started
These instructions will help you get the Atlas command-line tool up and running on your machine.

### Prerequisites
Please install the following dependencies before running the Atlas command-line tool.

#### protoc-gen-go
Protobuf generator for go

```sh
go get -u github.com/golang/protobuf/protoc-gen-go
```
For more details visit [protoc-gen-go](https://developers.google.com/protocol-buffers/docs/gotutorial).

#### dep

This is a dependency management tool for Go. You can install `dep` with Homebrew:

```sh
$ brew install dep
```
More detailed installation instructions are available on the [GitHub repository](https://github.com/golang/dep).

#### golang-migrate
Bootstrapped applications use [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations. You can install the `migrate` binary with the standard Go toolchain:

```
$ go get -u -d github.com/golang-migrate/migrate/cli github.com/lib/pq
$ go build -tags 'postgres' -o /usr/local/bin/migrate github.com/golang-migrate/migrate/cli
```
See the official golang-migrate [GitHub repository](https://github.com/golang-migrate/migrate) for more information about this tool.

### Installing
The following steps will install the `atlas` binary to your `$GOBIN` directory.

```sh
$ go get github.com/infobloxopen/atlas-cli/atlas
```
You're all set! Alternatively, you can clone the repository and install the binary manually.

```sh
$ git clone https://github.com/infobloxopen/atlas-cli.git
$ cd atlas-cli
$ make
```

## Bootstrap an Application
Rather than build applications completely from scratch, you can leverage the command-line tool to initialize a new project. This will generate the necessary files and folders to get started.

```sh
$ atlas init-app -name=my-application
$ cd my-application
```
#### Flags
Here's the full set of flags for the `init-app` command.

| Flag          | Description                                                         | Required      | Default Value |
| ------------- | ------------------------------------------------------------------- | ------------- | ------------- |
| `name`        | The name of the new application                                     | Yes           | N/A           |
| `db`          | Bootstrap the application with PostgreSQL database integration      | No            | `false`       |
| `gateway`     | Initialize the application with a gRPC gateway                      | No            | `false`       |
| `health`      | Initialize the application with internal health checks              | No            | `false`       |
| `pubsub`      | Initialize the application with a atlas-pubsub example              | No            | `false`       |
| `registry`    | The Docker registry where application images are pushed             | No            | `""`          |
| `create`      | Initialize the application with additional services based on a file | No            | `""`          |

You can run `atlas init-app --help` to see these flags and their descriptions on the command-line.

#### Additional Examples


```sh
# generates an application with a grpc gateway 
atlas init-app -name=my-application -gateway
```

```sh
# generates an application with a postgres database
atlas init-app -name=my-application -db
```

```sh
# specifies a docker registry
atlas init-app -name=my-application -registry=infoblox
```

```sh
# generates an application with additional services
# the input file (expand.txt in the example below) 
#   is a list of strings (letters only), one service name on each line
atlas init-app -name=my-application -expand=expand.txt -db=true
```

```sh
# example `expand.txt` file for use with -expand option
# Each line must either be a single string of letters only, 
# or two strings separated by a single comma. The latter option
# Is for the use case where the user wants a customized pluralization
# of a word. In the example below, "Artifacts" will pluralize to 
# "Artifactss", and "Kubernetes" will pluralize to "KubeCluster"
# In general, the best practice for object names should be 
# either <singular> or <singular,plural>

Artifacts
AwsInstance
Kubernetes,KubeCluster
```
Images names will vary depending on whether or not a Docker registry has been provided.

```sh
# docker registry was provided
registry-name/image-name:image-version
```

```sh
# docker registry was not provided
image-name:image-version
```

### Pubsub Example
To run the pubsub example ensure you run the application with the correct configuration by passing the pubsub server address. 
For more info  [atlas-pubsub](https://github.com/infobloxopen/atlas-pubsub)
```sh
# generates an application with a pubsub example
atlas init-app -name=my-application -pubsub
```

Of course, you may include all the flags in the `init-app` command.

```sh
atlas init-app -name=my-application -gateway -db -registry=infoblox -pubsub -health
```

## Viper Configuration

Generated atlas projects use [viper](https://github.com/spf13/viper), a complete configuration solution that allows an application to run from different environments. Viper also provides precedence order which is in the order as below.

#### Running from Default Values 
By default if you don't change anything your application will run with the values in config.go 
#### Running from Flags 
```
go run cmd/server/*.go --database.port 5432
```
#### Running from Environment Variables 
```
export DATABASE_PORT=5432
go run cmd/server/*.go
```
#### Running from Config file  
Change the configuration for defaultConfigDirectory and defaultConfigFile to point to your configuration file. 
You can either change it in config.go, passing it as environment variables, or flags. 

```
go run cmd/server/*.go --config.source "some/path/" --config.file "config_file.yaml" 
```


### Manually adding Viper 
1. Copy  [config.go](./atlas/config.gotmpl) and add it to your project under cmd/server/config.go
2. Update config.go by setting all your default values 
3. Add the following snippet of code inside your main.go that will allow you to initilize all the viper configuration. 
```go
func init() {
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AddConfigPath(viper.GetString("config.source"))
	if viper.GetString("config.file") != "" {
		log.Printf("Serving from configuration file: %s", viper.GetString("config.file"))
		viper.SetConfigName(viper.GetString("config.file"))
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("cannot load configuration: %v", err)
		}
	} else {
		log.Printf("Serving from default values, environment variables, and/or flags")
	}
	resource.RegisterApplication(viper.GetString("app.id"))
	resource.SetPlural()
}
```
4. To get or set viper configuration inside your code use the following methods:
```go
// Retrieving a string
viper.GetString("database.address")
// Retrieving a bool 
viper.GetBool("database.enable")
```




## Contributing to the Atlas CLI
Contributions to the Atlas CLI are welcome via pull requests and issues. If you're interested in making changes or adding new features, please take a minute to skim these instructions.

### Regenerating Template Bindata
The `templates/` directory contains a set of Go templates. When the Atlas CLI bootstraps a new application, it uses these templates to render new application files.

This project uses [go-bindata](https://github.com/jteeuwen/go-bindata) to package the Go templates and `atlas` source files together into a single binary. If your changes add or update templating, you need to regenerate the template bindata by running this command:

```sh
make templating
```

Your templating changes will take effect next time you run the `atlas` binary.

### Running the Integration Tests

The Atlas CLI integration tests ensure that new changes do not break exists features. To run the Atlas CLI unit and integration tests locally, set `e2e=true` in your environment.
```
make test-with-integration
```

### Adding New CLI Commands

If you wish to add a new command to the Atlas CLI, please take a look at the [command interface](https://github.com/infobloxopen/atlas-cli/blob/master/atlas/commands/command.go). This interface is intended to make adding new commands as minimally impactful to existing functionality.

To start out, consider looking at the [bootstrap command's implementation](https://github.com/infobloxopen/atlas-cli/blob/master/atlas/commands/bootstrap/bootstrap.go#L43) of this interface.
