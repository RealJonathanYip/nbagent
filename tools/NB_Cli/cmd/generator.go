package cmd

import (
	"fmt"
	"git.yayafish.com/nbagent/tools/NB_Cli/cmd/generator"
	"os"
	"path/filepath"

	//"git.yayafish.com/nbagent/log"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var (
	szBaseTemplatePath = filepath.Join(os.Getenv("GOPATH"), "src",
		"git.yayafish.com/nbagent/tools/NB_Cli/pb/template_file")

	szPBFileDescriptor = ""
	szTemplatesPath    = ""
	szOutputPath       = ""
)

func init() {
	RootCmd.AddCommand(objGeneratorCmd)
	objGeneratorCmd.Flags().StringVar(&szPBFileDescriptor, "pb-descriptor", "",
		"a FileDescriptorSet (a protocol buffer,defined in descriptor.proto) containing all of the input files to FILE")
	objGeneratorCmd.Flags().StringVar(&szTemplatesPath, "template-path", "",
		"template path")
	objGeneratorCmd.Flags().StringVar(&szOutputPath, "output-path", "./",
		"output path")

	if len(szTemplatesPath) == 0 {
		szTemplatesPath = szBaseTemplatePath
	}
}

var objGeneratorCmd = &cobra.Command{
	Use:   "gen",
	Short: "generator",

	Run: GeneratorFunc,
}

func GeneratorFunc(cmd *cobra.Command, args []string) {

	fmt.Printf("pb-descriptor: %v \n", szPBFileDescriptor)
	fmt.Printf("template-path: %v \n", szTemplatesPath)
	fmt.Printf("output path: %v \n", szOutputPath)

	var anyErr error
	var byteProto []byte
	byteProto, anyErr = ioutil.ReadFile(szPBFileDescriptor)
	if anyErr != nil {
		//log.Warningf("ioutil.ReadFile error: %v", anyErr)
		fmt.Printf("ioutil.ReadFile error: %v \n", anyErr)
		return
	}

	var objFileDescriptorSet = descriptor.FileDescriptorSet{}
	anyErr = proto.Unmarshal(byteProto, &objFileDescriptorSet)
	if anyErr != nil {
		fmt.Printf(" proto.Unmarshal error: %v \n", anyErr)
		return
	}
	fmt.Printf("FileDescriptorSet: %+v \n", objFileDescriptorSet)

	generateFile(&objFileDescriptorSet)

	return
}

func generateFile(ptrFileDescriptorSet *descriptor.FileDescriptorSet) {

	var objProtoBuilder = generator.NewProtoBuilder(szTemplatesPath, szOutputPath)
	anyErr := objProtoBuilder.ReadProtoFile(*ptrFileDescriptorSet.File[0])
	if anyErr != nil {
		return
	}

	var szOutFile = "/" + *ptrFileDescriptorSet.File[0].Package + ".nbagent.go"
	_ = objProtoBuilder.Build("/nbagent_pb.tmpl", szOutFile)
}
