package generator

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type ProtoBuilder struct {
	FileDescProto descriptor.FileDescriptorProto
	TemplatesPath string
	OutputPath    string
}

func NewProtoBuilder(szTemplatesPath, szOutputPath string) ProtoBuilder {
	return ProtoBuilder{
		TemplatesPath: szTemplatesPath,
		OutputPath:    szOutputPath,
	}
}

func (ptrBuilder *ProtoBuilder) ReadProtoFile(objFileProto descriptor.FileDescriptorProto) error {

	var objProto = objFileProto
	for _, ptrService := range objFileProto.Service {
		for _, ptrMethod := range ptrService.Method {
			var szInputType string = *ptrMethod.InputType
			aryInputType := strings.Split(szInputType, ".")
			if len(aryInputType) > 0 {
				szInputType = aryInputType[len(aryInputType)-1]
				ptrMethod.InputType = &szInputType
			}

			var szOutputType string = *ptrMethod.OutputType
			aryOutputType := strings.Split(szOutputType, ".")
			if len(aryOutputType) > 0 {
				szOutputType = aryOutputType[len(aryOutputType)-1]
				ptrMethod.OutputType = &szOutputType
			}
		}
	}

	ptrBuilder.FileDescProto = objProto
	return nil
}

func (ptrBuilder *ProtoBuilder) Build(szTemplateFile, szOutFile string) error {
	var anyErr error
	// get parse result in ptrBuffer
	ptrBuffer := new(bytes.Buffer)
	anyErr = ParseTemplate(ptrBuffer, path.Join(ptrBuilder.TemplatesPath, szTemplateFile), ptrBuilder.FileDescProto)
	if anyErr != nil {
		fmt.Printf("ParseTemplate error: %v \n", anyErr)
		return anyErr
	}
	// ptrBuffer to file
	szOutFilePath := path.Join(ptrBuilder.OutputPath, szOutFile)
	_, anyErr = os.Create(szOutFilePath)
	if anyErr != nil {
		fmt.Printf("os.Create error: %v \n", anyErr)
		return anyErr
	}

	anyErr = ioutil.WriteFile(szOutFilePath, ptrBuffer.Bytes(), 0644)
	if anyErr != nil {
		fmt.Printf("ioutil.WriteFile error: %v \n", anyErr)
		return anyErr
	}
	return nil
}

func ParseTemplate(ptrBuffer *bytes.Buffer, szTemplatePath string, objData interface{}) error {

	ptrTemplate := getTemplate(szTemplatePath)
	if ptrTemplate == nil {
		return fmt.Errorf("parse file error")
	}
	anyErr := ptrTemplate.Execute(ptrBuffer, objData)
	if anyErr != nil {
		fmt.Printf("Template.Execute error: %v \n", anyErr)
		return anyErr
	}
	return nil
}

func getTemplate(templatePath string) *template.Template {

	ptrTemplate, anyErr := template.New(filepath.Base(templatePath)).Funcs(getFuncMap()).ParseFiles(templatePath)
	if anyErr != nil {
		fmt.Printf("template.New error: %v \n", anyErr)
		return nil
	}
	return ptrTemplate
}

func getFuncMap() template.FuncMap {
	return template.FuncMap{
		"firstUpperCase": strings.Title,
		// don't firstLower, temp use ToLower
		"loweCase": strings.ToLower,
	}
}
