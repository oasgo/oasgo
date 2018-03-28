example-client:
	go run *.go generate client -f testdata/pets.yaml | goimports | gofmt > example/client/client.go
example-server:
	go run *.go generate server -f testdata/pets.yaml | goimports | gofmt > example/server/server.go
example-test: example-client
	go test -race -v ./example/client
