install:
	@go install
example-client: install
	@oasgo generate client -f testdata/pets.yaml | goimports  > example/client/client.go
example-handlers: install
	@oasgo generate handlers -f testdata/pets.yaml | goimports > example/server/handlers.go
example-handlers-v2: install
	@oasgo generate handlers-v2 -f testdata/pets.yaml | goimports > example/server/handlersv2.go
example-test: example-client
	go test -race -v ./example/client
