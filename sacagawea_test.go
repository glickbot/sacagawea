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

package sacagawea_test

import (
	"encoding/json"
	"github.com/glickbot/sacagawea"
	"testing"
)

const discoveryUrl = "https://www.googleapis.com/discovery/v1/apis/discovery/v1/rest"
const cloudfunctionsUrl = "https://cloudfunctions.googleapis.com/$discovery/rest?version=v1"
func TestCloudFunctionsClient(t *testing.T){
	renderer, err := sacagawea.NewServiceRenderer(
		cloudfunctionsUrl,
		"cloudfunctions",
		"./cloudfunctions-v1",
	)
	if err != nil {
		t.Error(err)
	}
	if err := renderer.RenderAll(); err != nil {
		t.Error(err)
	}
}

func TestServiceRenderer(t *testing.T){
	renderer, err := sacagawea.NewServiceRenderer(
		discoveryUrl,
		"discovery",
		"./discovery",
		)
	if err != nil {
		t.Error(err)
	}
	if err := renderer.RenderAll(); err != nil {
		t.Error(err)
	}
}


func TestListServices(t *testing.T){
	list, err := sacagawea.ListServices()
	if err != nil {
		t.Error(err)
	}
	apis, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		t.Error(err)
	}
	_, err = UnmarshalTestAPIList(apis)
	if err != nil {
		t.Error(err)
	}
	//fmt.Printf("%s\n", apis)
}

func UnmarshalTestAPIList(data []byte) (TestAPIList, error) {
	var r TestAPIList
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TestAPIList) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type TestAPIList struct {
	Kind             string `json:"kind"`
	DiscoveryVersion string `json:"discoveryVersion"`
	Items            []Item `json:"items"`
}

type Item struct {
	Kind              Kind    `json:"kind"`
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Version           string  `json:"version"`
	Title             string  `json:"title"`
	Description       string  `json:"description"`
	DiscoveryRESTURL  string  `json:"discoveryRestUrl"`
	DiscoveryLink     string  `json:"discoveryLink"`
	Icons             Icons   `json:"icons"`
	DocumentationLink string  `json:"documentationLink"`
	Labels            []Label `json:"labels"`
	Preferred         bool    `json:"preferred"`
}

type Icons struct {
	X16 string `json:"x16"`
	X32 string `json:"x32"`
}

type Kind string
const (
	DiscoveryDirectoryItem Kind = "discovery#directoryItem"
)

type Label string
const (
	Deprecated Label = "deprecated"
	Labs Label = "labs"
	LimitedAvailability Label = "limited_availability"
)
