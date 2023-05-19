build-options:
	buf generate --template proto/options/buf.gen.yaml --path proto/options
build-example:
	go install
	go install github.com/favadi/protoc-go-inject-tag@latest
	cd example;	buf generate; protoc-go-inject-tag -input *.go