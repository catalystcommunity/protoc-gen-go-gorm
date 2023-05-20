build-options:
	cd options; buf generate
build-example:
	go install
	go install github.com/favadi/protoc-go-inject-tag@latest
	cd example;	buf generate; protoc-go-inject-tag -input *.go