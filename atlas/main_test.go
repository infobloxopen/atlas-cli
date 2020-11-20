package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func e2eTeardown() {
	log.Print("cleaning up bootstrapped files")
	if err := os.RemoveAll("test"); err != nil {
		log.Fatalf("failed to delete test folder: %v", err)
	}
}

func TestMain(m *testing.M) {
	if len(os.Getenv("e2e")) == 0 {
		log.Print("skipping end-to-end tests")
		return
	}

	if err := os.RemoveAll("test"); err != nil {
		log.Fatalf("failed to delete test folder: %v", err)
	}
	log.Print("installing atlas cli")
	if out, err := exec.Command("go", "install").CombinedOutput(); err != nil {
		log.Print(string(out))
		log.Fatalf("failed to install atlas cli: %v", err)
	}
	log.Print("running init-app")
	if out, err := exec.Command("atlas", "init-app", "-name=test", "-gateway", "-health", "-helm", "-pubsub", "-debug").CombinedOutput(); err != nil {
		log.Print(string(out))
		log.Fatalf("failed to run atlas init-app: %v", err)
	}
	defer e2eTeardown()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	output := "bin/server"
	log.Printf("building server")
	build := exec.Command("go", "build", "-mod", "vendor", "-o", output, "./cmd/server")
	basePath := fmt.Sprintf("%s/test", dir)
	build.Dir = basePath
	//go build -o  ./test/bin/server ./test/cmd/server
	if out, err := build.CombinedOutput(); err != nil {
		log.Print(string(out))
		log.Fatalf("failed to build server: %v", err)
	}
	log.Printf("runnning server")
	cmd := exec.CommandContext(ctx, fmt.Sprintf("%s/%s", basePath, output))
	stderr, _ := cmd.StderrPipe()
	cmdLog := bufio.NewScanner(stderr)
	ready := make(chan struct{}, 1)
	go func() {
		first := true
		for cmdLog.Scan() {
			line := cmdLog.Text()
			if first && strings.Contains(line, "serving") {
				ready <- struct{}{}
				first = false
			}
			log.Print(line)
		}
	}()
	if err := cmd.Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			log.Fatal()
		}
	}()
	log.Print("wait for servers to load up")
	<-ready
	time.Sleep(time.Second)

	code := m.Run()
	// os.Exit() does not respect defer statements
	e2eTeardown()
	os.Exit(code)
}

func TestGetVersion(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/test/v1/version")
	if err != nil {
		t.Errorf("expected get/version to succeed, but got error: %v", err)
	} else if resp.StatusCode != 200 {
		t.Errorf("expected response to be status 200, but got %d: %v", resp.StatusCode, resp)
	}
}

func TestFormatting(t *testing.T) {
	cmd := exec.Command("go", "fmt", "./...")
	cmd.Dir = "./test"
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("unable to run go fmt: %v", err)
	}
	// check if bootstrap command produced unformatted go code
	if string(out) != "" {
		// print unformatted files on a single line, not multiple lines
		files := strings.Split(string(out), "\n")
		t.Fatalf("test application has unformatted go code: %v", strings.Join(files, " "))
	}
}

func TestHelmLint(t *testing.T) {
	cmd := exec.Command("helm", "lint", "helm/test")
	cmd.Dir = "./test"
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helm lint failed: %v\n%s", err, string(out))
	}
}
