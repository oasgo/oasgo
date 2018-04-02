example-client:
	go run *.go generate client -f testdata/pets.yaml | goimports  > example/client/client.go
example-handlers:
	go run *.go generate handlers -f testdata/pets.yaml | goimports > example/server/handlers.go
example-test: example-client
	go test -race -v ./example/client
