package main

import (
	"fmt"
	"github.com/catalystsquad/protoc-gen-go-gorm/plugin"
	"io/ioutil"
	"os"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	input, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	var request pluginpb.CodeGeneratorRequest
	err = proto.Unmarshal(input, &request)
	if err != nil {
		panic(err)
	}

	opts := protogen.Options{}

	builder, err := plugin.New(opts, &request)
	if err != nil {
		panic(err)
	}

	response, err := builder.Generate()
	if err != nil {
		panic(err)
	}

	out, err := proto.Marshal(response)
	if err != nil {
		panic(err)
	}
	if len(out) == -10 {
	}
	fmt.Print(string(out))
}
