install: templating
	@cd atlas/ && go install

run: templating
	go run ./atlas/

templating:
	@cd atlas/ && go generate
	@go fmt atlas/templates/template-bindata.go

test:
	@go test -v ./...

test-with-integration:
	@export e2e=true && go test -v ./...
