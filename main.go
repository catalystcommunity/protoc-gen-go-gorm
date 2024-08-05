package main

import (
	"flag"
	"fmt"
	"github.com/catalystcommunity/app-utils-go/env"
	gorm "github.com/catalystcommunity/protoc-gen-go-gorm/options"
	"github.com/catalystcommunity/protoc-gen-go-gorm/plugin"
	"github.com/golang/glog"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"reflect"
)

var logLevel = env.GetEnvOrDefault("LOG_LEVEL", "ERROR")

func main() {
	flag.Set("stderrthreshold", logLevel)
	flag.Parse()
	defer glog.Flush()
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gp *protogen.Plugin) error {
		gp.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		for _, name := range gp.Request.FileToGenerate {
			f := gp.FilesByPath[name]

			if len(f.Messages) == 0 {
				glog.Infof("Skipping %s, no messages", name)
				continue
			}

			glog.Infof("Processing %s", name)
			glog.Infof("Generating %s\n", fmt.Sprintf("%s.pb.gorm.go", f.GeneratedFilenamePrefix))

			if shouldGenerateFile(f) {
				gf := gp.NewGeneratedFile(fmt.Sprintf("%s.pb.gorm.go", f.GeneratedFilenamePrefix), f.GoImportPath)
				err := plugin.ApplyTemplate(gf, f)
				if err != nil {
					gf.Skip()
					gp.Error(err)
					continue
				}
			}

		}

		return nil
	})
}

func shouldGenerateFile(file *protogen.File) bool {
	options := getFileOptions(file)
	return options != nil && options.Generate
}

func getFileOptions(file *protogen.File) *gorm.GormFileOptions {
	options := file.Desc.Options().(*descriptorpb.FileOptions)
	if options == nil {
		return &gorm.GormFileOptions{}
	}
	v := proto.GetExtension(options, gorm.E_FileOpts)
	if reflect.ValueOf(v).IsNil() {
		return nil
	}
	opts, ok := v.(*gorm.GormFileOptions)
	if !ok {
		return nil
	}
	return opts
}
