package cloudfunctions

import (
	"encoding/json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"fmt"
	"os"
	"testing"
)

func TestCloudFunctionClient(t *testing.T) {
	credentials := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentials == "" {
		t.Error("Please set the environment variable GOOGLE_APPLICATION_CREDENTIALS to your credentials json")
	}
	project := os.Getenv("PROJECT_ID")
	if project == "" {
		t.Error("Please set the environment variable PROJECT_ID to a project (ideally with a cloud function)")
	}

	data, err := ioutil.ReadFile(credentials)
	if err != nil {
		t.Fatal(err)
	}
	conf, err := google.JWTConfigFromJSON(data, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		t.Fatal(err)
	}
	client := conf.Client(oauth2.NoContext)
	cf := NewClient("https://cloudfunctions.googleapis.com", client)

	locations, err := cf.Cloudfunctions_Projects_Locations_List("", fmt.Sprintf("projects/%s", project), 10, "")
	if err != nil {
		t.Fatalf("Locations API Error: %v\n", err)
	}

	for _, l := range locations.Default.Locations {
		// printJson(l)
		// continue
		funs, err := cf.Cloudfunctions_Projects_Locations_Functions_List(10, "", l.Name)
		if err != nil {
			t.Fatalf("API Error: %v\n", err)
		}

		for _, f := range funs.Default.Functions {
			printJson(f)
			//t.Printf("%d: %v\n", i, f)
		}
	}
}

func printJson(v interface{}){
	bytes, err := json.MarshalIndent(v, "", "   ")
	if err != nil {
		fmt.Printf("Error marshaling %v\n", v)
	}
	fmt.Printf("%s\n", bytes)
}