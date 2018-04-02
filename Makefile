example-client:
	go run *.go generate client -f testdata/pets.yaml | goimports  > example/client/client.go
example-server:
	go run *.go generate server -f testdata/pets.yaml | goimports > example/server/server.go
example-test: example-client
	go test -race -v ./example/client
