package bootstrap

import (
	"bufio"
	"fmt"
	"github.com/jinzhu/inflection"
	"io"
	"log"
	"os"
	"path"
	"strings"
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
	WithDatabase bool
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
		migrateStr = fmt.Sprintf("%05d", migrateNum)
		resource := templateResource {
			NameCamel: strcase.ToCamel(name),
			NameCamels: strPlural(strcase.ToCamel(name)),
			NameLowerCamel: strcase.ToLowerCamel(name),
			NameLowerCamels: strPlural(strcase.ToLowerCamel(name)),
			NameSnake: strcase.ToSnake(name),
			NameSnakes: strPlural(strcase.ToSnake(name)),
			MigrateVer: migrateStr,
			WithDatabase: withDatabase,
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
		line := scanner.Text()
		/*if strings.Contains(line, ",") {
			var plural = strings.Split(line, ",")
			if len(plural) != 2 {
				fmt.Printf("Lines containing a comma must have exactly two words with only letters, separated by one comma. Ignoring this line: %q", line)
			} else { // length is exactly 2
				if IsLetter(plural[0]) && IsLetter(plural[1]) {
					lines = append(lines, plural[0])
					inflection.AddIrregular(plural[0], plural[1])
				} else {
					fmt.Printf("Lines containing a comma must have exactly two words with only letters, separated by one comma. Ignoring this line: %q", line)
				}
			}
		} else if len(line) > 0 && IsLetter(line) {
			lines = append(lines, line)
		} else {
			fmt.Printf("Ignoring resource in config file with value: %q. Resource must be a single word with only letters.\r\n", line)
		}*/
		var words = strings.Split(line, ",")
		if len(words) == 1 && IsLetter(words[0]) {
			lines = append(lines, words[0])
		} else if len(words) == 2 && IsLetter(words[0]) && IsLetter(words[1]) {
			lines = append(lines, words[0])
			inflection.AddIrregular(words[0], words[1])
		} else {
			fmt.Printf("Ignoring resource in config file with value: %q. Resource must be a single word with only letters, or exactly two words separated by a comma.\r\n", line)
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
	// For use cases with words that end in s, user may add unique pluralization
	// If they don't want the plural of 'Artifacts' for example to be 'Artifactss'
	// In their input file for the -expand argument they may add a comma-separated line
	// The first word is the singular form, the second word is the plural form
	// If no word is given, the default "+s" method is used

	plural := inflection.Plural(s)
	if s == plural {
		return s + "s"
	} else {
		return plural
	}
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
