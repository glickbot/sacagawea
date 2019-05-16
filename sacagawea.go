// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sacagawea

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/googleapis/gnostic/OpenAPIv3"
	"github.com/googleapis/gnostic/compiler"
	discovery "github.com/googleapis/gnostic/discovery"
	plugins "github.com/googleapis/gnostic/plugins"
	surface "github.com/googleapis/gnostic/surface"
	"github.com/glickbot/gnostic/plugins/gnostic-go-generator/gorenderer"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func NewServiceRenderer(discoveryRestURL string, packageName string, output string) (*ServiceRenderer, error) {
	serviceRenderer := &ServiceRenderer{
		discoveryRestURL: discoveryRestURL,
		packageName: packageName,
		output: output,
	}
	err := serviceRenderer.load()
	if err != nil {
		return nil, err
	}
	serviceRenderer.renderer.Package = serviceRenderer.packageName
	return serviceRenderer, err
}

type ServiceRenderer struct {
	discoveryRestURL string
	packageName string
	output string
	renderer *gorenderer.Renderer
	model *surface.Model
}

func (s *ServiceRenderer) load() error {
	apiBytes, err := compiler.FetchFile(s.discoveryRestURL)
	if err != nil {
		return errors.New(fmt.Sprintf("Error while fetching URL: %v\n", err))
	}

	info, err := compiler.ReadInfoFromBytes("", apiBytes)
	if err != nil {
		return err
	}
	m, ok := compiler.UnpackMap(info)
	if !ok {
		log.Printf("%s", string(apiBytes))
		return errors.New("Invalid input")
	}
	document, err := discovery.NewDocument(m, compiler.NewContext("$root", nil))
	documentV3, err := OpenAPIv3(document)
	if err != nil {
		return err
	}
	s.model, err = surface.NewModelFromOpenAPI3(documentV3)
	if err != nil {
		return err
	}
	gorenderer.NewGoLanguageModel().Prepare(s.model)
	//modelJSON, _ := json.MarshalIndent(surfaceModel, "", "  ")
	s.renderer, err = gorenderer.NewServiceRenderer(s.model)
	if err != nil {
		return err
	}
	return nil
}

func (s *ServiceRenderer) RenderAll() error {
	if err := s.RenderClient(""); err != nil {
		return err
	}
	if err := s.RenderTypes(""); err != nil {
		return err
	}
	if err := s.RenderConstants(""); err != nil {
		return err
	}
	if err := s.RenderModel(""); err != nil {
		return err
	}
	if err := s.RenderProvider(""); err != nil {
		return err
	}
	if err := s.RenderServer(""); err != nil {
		return err
	}
	return nil
}

func (s *ServiceRenderer) RenderClient(filename string) error {
	if filename == "" {
		filename = "client.go"
	}
	renderedBytes, err := s.renderer.RenderClient()
	if err != nil {
		return err
	}
	return s.saveFile(renderedBytes, filename)
}

func (s *ServiceRenderer) RenderTypes(filename string) error {
	if filename == "" {
		filename = "types.go"
	}
	renderedBytes, err := s.renderer.RenderTypes()
	if err != nil {
		return err
	}
	return s.saveFile(renderedBytes, filename)
}
func (s *ServiceRenderer) RenderConstants(filename string) error {
	if filename == "" {
		filename = "constants.go"
	}
	renderedBytes, err := s.renderer.RenderConstants()
	if err != nil {
		return err
	}
	return s.saveFile(renderedBytes, filename)
}

func (s *ServiceRenderer) RenderModel(filename string) error {
	if filename == "" {
		filename = "model.json"
	}
	renderedBytes, err := json.MarshalIndent(s.model, "", "  ")
	if err != nil {
		return err
	}
	return s.saveFile(renderedBytes, filename)
}

func (s *ServiceRenderer) RenderProvider(filename string) error {
	if filename == "" {
		filename = "provider.go"
	}
	renderedBytes, err := s.renderer.RenderProvider()
	if err != nil {
		return err
	}
	return s.saveFile(renderedBytes, filename)
}

func (s *ServiceRenderer) RenderServer(filename string) error {
	if filename == "" {
		filename = "server.go"
	}
	renderedBytes, err := s.renderer.RenderServer()
	if err != nil {
		return err
	}
	return s.saveFile(renderedBytes, filename)
}


func (s *ServiceRenderer) saveFile(fileBytes []byte, filename string) error {
	if filepath.Ext(filename) == ".go" {
		var err error
		fileBytes, err = imports.Process(filename, fileBytes, nil)
		if err != nil {
			return err
		}
	}
	return ioutil.WriteFile(
		fmt.Sprintf("%s/%s", s.output, filename),
		fileBytes,
		0644)
}



func ListServices() (*List, error) {
	listBytes, err := compiler.FetchFile(APIsListServiceURL)
	if err != nil {
		return nil, err
	}
	// Unpack the apis/list response
	return NewList(listBytes)
}

func GenServiceClient(discoveryRestURL string, path string) error {

	apiBytes, err := compiler.FetchFile(discoveryRestURL)
	if err != nil {
		return errors.New(fmt.Sprintf("Error while fetching URL: %v\n", err))
	}
	// Export any requested formats.
	return exportBytes(apiBytes, path)
}

func GetFiles(apiBytes []byte, path string) error {
	// Unpack the discovery document.
	info, err := compiler.ReadInfoFromBytes("", apiBytes)
	if err != nil {
		return err
	}
	m, ok := compiler.UnpackMap(info)
	if !ok {
		log.Printf("%s", string(apiBytes))
		return errors.New("Invalid input")
	}
	document, err := discovery.NewDocument(m, compiler.NewContext("$root", nil))
	documentV3, err := OpenAPIv3(document)
	if err != nil {
		return err
	}
	surfaceModel, err := surface.NewModelFromOpenAPI3(documentV3)
	if err != nil {
		return err
	}
	gorenderer.NewGoLanguageModel().Prepare(surfaceModel)
	//modelJSON, _ := json.MarshalIndent(surfaceModel, "", "  ")
	renderer, err := gorenderer.NewServiceRenderer(surfaceModel)
	if err != nil {
		return err
	}
	typeBytes, err := renderer.RenderTypes()
	if err != nil {
		return err
	}
	fmt.Printf("%s", typeBytes)
	return nil
}


func exportBytes(apiBytes []byte, path string) error {
	// Unpack the discovery document.
	info, err := compiler.ReadInfoFromBytes("", apiBytes)
	if err != nil {
		return err
	}
	m, ok := compiler.UnpackMap(info)
	if !ok {
		log.Printf("%s", string(apiBytes))
		return errors.New("Invalid input")
	}
	document, err := discovery.NewDocument(m, compiler.NewContext("$root", nil))

	// Generate the OpenAPI 3 equivalent.
	openAPIDocument, err := OpenAPIv3(document)
	if err != nil {
		return err
	}
	apiBytes, err = proto.Marshal(openAPIDocument)
	if err != nil {
		return err
	}
	//filename := "openapi3-" + document.Name + "-" + document.Version + ".pb"
	//err = ioutil.WriteFile(filename, apiBytes, 0644)
	//if err != nil {
	//	return err
	//}
	request := &plugins.Request{}
	documentV3 := &openapi_v3.Document{}
	err = proto.Unmarshal(apiBytes, documentV3)

	request.AddModel("openapi.v3.Document", documentV3)
	surfaceModel, err := surface.NewModelFromOpenAPI3(documentV3)
	if err == nil {
		request.AddModel("surface.v1.Model", surfaceModel)
	}

	requestBytes, _ := proto.Marshal(request)

	cmd := exec.Command("gnostic-go-generator", "-plugin")
	cmd.Stdin = bytes.NewReader(requestBytes)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	response := &plugins.Response{}
	err = proto.Unmarshal(output, response)

	if err != nil {
		return errors.New("Invalid plugin response (plugins must write log messages to stderr, not stdout).")
	}
	// fmt.Printf("%s\n", spew.Sdump(response));
	//
	//if err != nil {
	//	return err
	//}

	err = plugins.HandleResponse(response, path)
	return err
}