install:
	@go install
example-client: install
	@oasgo generate client -f testdata/pets.yaml | goimports  > example/client/client.go
example-dto: install
	@oasgo generate dto -f testdata/pets.yaml | goimports > example/server/handlers.go
example-test: example-client
	go test -race -v ./example/client
