example-client:
	go run *.go generate -f testdata/pets.yaml | goimports | gofmt > example/client/client.go
example-test: example-client
	go test -race -v ./example/client
