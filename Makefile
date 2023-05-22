options:
	buf generate --template proto/options/buf.gen.yaml --path proto/options
build-options:
	buf generate --template proto/options/buf.gen.yaml --path proto/options
build-example:
	go install
	go install github.com/favadi/protoc-go-inject-tag@latest
	buf generate --template example/demo/buf.gen.yaml --path example/demo
	protoc-go-inject-tag -input example/demo/*.*.*.go
	protoc-go-inject-tag -input example/demo/*.*.go