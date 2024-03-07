install: templating
	@cd atlas/ && go install

run: templating
	go run ./atlas/

.bindata:
	go install github.com/go-bindata/go-bindata/v3/go-bindata@v3
	touch $@

templating: .bindata
	@cd atlas/ && rm -f templates/template-bindata.go && go generate && go fmt templates/template-bindata.go

test:
	@go test -v ./...

test-with-integration: templating
	@export e2e=true && go test -v ./... -count=1
