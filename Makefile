install:
	@cd atlas/ && go install

templating:
	@cd atlas/commands/bootstrap && go generate
	@go fmt atlas/commands/bootstrap/templates/template-bindata.go

test:
	@go test -v ./...

test-with-integration:
	@export e2e=true && go test -v ./...
