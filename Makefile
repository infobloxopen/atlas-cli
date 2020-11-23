install: templating
	@cd atlas/ && go install

run: templating
	go run ./atlas/

.bindata:
	go get -u github.com/go-bindata/go-bindata/...
	touch $@

templating: .bindata
	@cd atlas/ && rm -f templates/template-bindata.go && go generate
	@go fmt atlas/templates/template-bindata.go

test:
	@go test -v ./...

test-with-integration: templating
	@export e2e=true && go test -v ./... -count=1
