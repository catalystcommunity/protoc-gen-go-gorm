build-options:
	buf generate --template proto/options/buf.gen.yaml --path proto/options
build-example:
	go install
	go install github.com/favadi/protoc-go-inject-tag@latest
	go install github.com/mitchellh/protoc-gen-go-json@latest
	buf generate --template example/cockroachdb/buf.gen.yaml --path example/cockroachdb
	protoc-go-inject-tag -input example/cockroachdb/*.*.*.go
	protoc-go-inject-tag -input example/cockroachdb/*.*.go
	buf generate --template example/postgres/buf.gen.yaml --path example/postgres
	protoc-go-inject-tag -input example/postgres/*.*.*.go
	protoc-go-inject-tag -input example/postgres/*.*.go
clean:
	rm -f example/cockroachdb/*.go
	rm -f example/postgres/*.go
	rm -f options/*.go
generate: clean build-options build-example
test: generate
	go test -v ./test