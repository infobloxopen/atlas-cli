package bootstrap

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"text/template"

	"github.com/iancoleman/strcase"
)

type finalTemplate struct {
	AppName string
	WithDatabase bool
	R []templateResource
}

type templateResource struct {
	NameCamel string
	NameCamels string
	NameLowerCamel string
	NameLowerCamels string
	NameSnake string
	NameSnakes string
	MigrateVer string
	Path string
}

func expandResource(appName, expandName string, withDatabase bool) error {

	lines, err := readLines("../" + expandName)
	if err != nil {
		panic(err)
	}

	r := []templateResource{}

	migrateNum := 2
	migrateStr := ""

	for _, name := range lines {
		if (migrateNum > 9) {
			migrateStr = "000" + strconv.Itoa(migrateNum)
		} else {
			migrateStr = "0000" + strconv.Itoa(migrateNum)
		}
		resource := templateResource {
			NameCamel: strcase.ToCamel(name),
			NameCamels: strPlural(strcase.ToCamel(name)),
			NameLowerCamel: strcase.ToLowerCamel(name),
			NameLowerCamels: strPlural(strcase.ToLowerCamel(name)),
			NameSnake: strcase.ToSnake(name),
			NameSnakes: strPlural(strcase.ToSnake(name)),
			MigrateVer: migrateStr,
		}
		r = append(r, resource)
		migrateNum += 1
	}

	err = runTemplate(r, appName, withDatabase,
		"../atlas/templates/pkg/pb/template.proto.gotmpl",
		"pkg/pb/" + appName + ".proto" )

	if err != nil {
		log.Fatalf("failed to create pkg/pb/" + appName + ".proto\n%s\n", err)
	}

	err = runTemplate(r, appName, withDatabase,
		"../atlas/templates/pkg/svc/servers.gotmpl",
		"pkg/svc/servers.go" )

	if err != nil {
		log.Fatalf("failed to create pkg/pb/servers.go\n%s\n", err)
	}

	err = runTemplate(r, appName, withDatabase,
		"../atlas/templates/cmd/server/endpoints.gotmpl",
		"cmd/server/endpoints.go" )

	if err != nil {
		log.Fatalf("failed to create cmd/server/endpoints.go\n%s\n", err)
	}

	err = runTemplate(r, appName, withDatabase,
		"../atlas/templates/cmd/server/servers.gotmpl",
		"cmd/server/servers.go" )

	if err != nil {
		log.Fatalf("failed to create cmd/server/servers.go\n%s\n", err)
	}

	os.MkdirAll("db/migration", os.ModePerm)

	for _ , res := range r {
		err = runTemplate([]templateResource{res}, appName, withDatabase,
			"../atlas/templates/db/migration/down.sql.gotmpl",
			"db/migration/" + res.MigrateVer + "_" + res.NameSnakes +  ".down.sql" )

		if err != nil {
			log.Fatalf("failed to create db/migration/" + res.MigrateVer + "_" + res.NameSnakes +  ".down.sql\n%s\n", err)
		}

		err = runTemplate([]templateResource{res}, appName, withDatabase,
			"../atlas/templates/db/migration/up.sql.gotmpl",
			"db/migration/" + res.MigrateVer + "_" + res.NameSnakes +  ".up.sql" )

		if err != nil {
			log.Fatalf("failed to create db/migration/" + res.MigrateVer + "_" + res.NameSnakes +  ".up.sql\n%s\n", err)
		}
	}

	return nil
}

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if len(scanner.Text()) > 0 && IsLetter(scanner.Text()) {
			lines = append(lines, scanner.Text())
		} else {
			fmt.Println("Ignoring resource in config file with value: <" + scanner.Text() + ">. Resource must be a single word with only letters.")
		}
	}
	return lines, scanner.Err()
}

func IsLetter(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') {
			return false
		}
	}
	return true
}

func strPlural(s string) string {
	// Had some use cases like AwsRds or Kubernetes and tried to
	// fix the plural since plural is not right like: Kubernetess
	// This is fine for URLs or Table names but breaks proto since
	// message and service definitions have collision:
	// sc-l-seizadi:cmdb seizadi$ make protobuf
	// github.com/seizadi/cmdb/pkg/pb/cmdb.proto:144:9: "AwsRds" is already defined in "api.cmdb".
	// github.com/seizadi/cmdb/pkg/pb/cmdb.proto:1087:9: "Kubernetes" is already defined in "api.cmdb".
	// So we need to have people pick names that have natural plural so in the above case
	// AwsRds -> AwsRdsInstance or Kubernetes => KubeCluster
	//if (s[len(s)-1:] == "s") {
	//	return s
	//}
	return s + "s"
}


func runTemplate(r []templateResource, appName string, expandName bool, src string, dst string ) error {
	// Create a new template and parse the file into it
	name := path.Base(src)
	t, err := template.New(name).ParseFiles(src)
	if err != nil {
		log.Fatalf("parsing template: %s\n", err)
	}
	// Create Template
	f, err := os.Create(dst)
	if err != nil {
		log.Fatalf("create file %s failed: %s\n", dst, err)
	}
	q := finalTemplate{appName, expandName,r}
	err = t.Execute(f, q)
	if err != nil {
		log.Fatalf("failed executing template: %s\n", err)
	}

	return nil
}

//Appends the contents of fileTwo to the end of fileOne, then deletes fileTwo
func CombineFiles(fileOne, fileTwo string) error {

	in, err := os.Open(fileTwo)
	if err != nil {
		log.Fatalln("failed to open second file for reading:", err)
	}
	defer in.Close()

	out, err := os.OpenFile(fileOne, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln("failed to open first file for writing:", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		log.Fatalln("failed to append second file to first:", err)
	}

	// Delete the old input file
	in.Close()
	out.Close()

	if err := os.Remove(fileTwo); err != nil {
		log.Fatalln("failed to remove", fileTwo)
	}

	return nil
}
