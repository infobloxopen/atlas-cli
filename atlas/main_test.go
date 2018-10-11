package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

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
	if out, err := exec.Command("atlas", "init-app", "-name=test", "-gateway", "-health", "-pubsub").CombinedOutput(); err != nil {
		log.Print(string(out))
		log.Fatalf("failed to run atlas init-app: %v", err)
	}
	defer func() {
		log.Print("cleaning up bootstrapped files")
		if err := os.RemoveAll("test"); err != nil {
			log.Fatalf("failed to delete test folder: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	output := "./test/bin/server"
	log.Printf("building server")
	build := exec.Command("go", "build", "-o", output, "./test/cmd/server")
	if out, err := build.CombinedOutput(); err != nil {
		log.Print(string(out))
		log.Fatalf("failed to build server: %v", err)
	}
	log.Printf("runnning server")
	if err := exec.CommandContext(ctx, output).Start(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
	log.Print("wait for servers to load up")
	time.Sleep(time.Second)

	m.Run()
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
