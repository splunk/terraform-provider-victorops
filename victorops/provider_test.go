package victorops

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/victorops/go-victorops/victorops"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider
var tfTemplate *template.Template

func init() {
	tfTemplate = template.Must(tfTemplate.ParseGlob("../templates/*.tf"))
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"victorops_config_test": testAccProvider,
		"victorops": testAccProvider,
	}
}

func getTestTemplate(f string, d interface{}) string {
	buf := &bytes.Buffer{}
	err := tfTemplate.ExecuteTemplate(buf, f, d)
	if err != nil {
		log.Fatalln(err)
		log.Fatalln("Could not find template file", f)
	}
	return buf.String()
}


func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("VO_API_ID"); v == "" {
		t.Fatal("VO_API_ID must be set for acceptance tests")
	}
	if v := os.Getenv("VO_API_KEY"); v == "" {
		t.Fatal("VO_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("VO_BASE_URL"); v == "" {
		t.Fatal("VO_BASE_URL must be set for acceptance tests")
	}
	if v := os.Getenv("VO_REPLACEMENT_USERNAME"); v == "" {
		t.Fatal("VO_REPLACEMENT_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("VO_ROTATION_GROUP_SLUG"); v == "" {
		t.Fatal("VO_ROTATION_GROUP_SLUG must be set for acceptance tests")
	}
}

func createTempConfigFile(content string, name string) (*os.File, error) {
	tempFile, err := ioutil.TempFile(os.TempDir(), name)
	if err != nil {
		return nil, fmt.Errorf("error creating temporary test file. err: %s", err.Error())
	}

	_, err = tempFile.WriteString(content)
	if err != nil {
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("error writing to temporary test file. err: %s", err.Error())
	}

	return tempFile, nil
}

// Create a new test client for unit and Acceptance tests
func newTestingClient() (*victorops.Client, error) {
	testClientAPIId, exists := os.LookupEnv("VO_TF_API_ID")
	if !exists {
		return nil, fmt.Errorf("error: required environment variable 'VO_TF_API_ID' is not set")
	}

	testClientAPIkey, exists := os.LookupEnv("VO_TF_API_KEY")
	if !exists {
		return nil, fmt.Errorf("error: required environment variable 'VO_TF_API_KEY' is not set")
	}

	TestClientBaseURL, exists := os.LookupEnv("VO_TF_BASE_URL")
	if !exists {
		TestClientBaseURL = "https://api.victorops.com"
	}

	client := victorops.NewClient(testClientAPIId, testClientAPIkey, TestClientBaseURL)
	return client, nil
}

func TestNewTestingClient(t *testing.T) {

	old := os.Getenv("VO_TF_API_ID")
	os.Setenv("VO_TF_API_ID", "1234")
	defer os.Setenv("VO_TF_API_ID", old)

	old = os.Getenv("VO_TF_API_KEY")
	os.Unsetenv("VO_TF_API_KEY")
	defer os.Setenv("VO_TF_API_KEY", old)

	old = os.Getenv("VO_TF_BASE_URL")
	os.Setenv("VO_TF_BASE_URL", "https://api.victorops.com")
	defer os.Setenv("VO_TF_BASE_URL", old)

	if client, err := newTestingClient(); client != nil {
		t.Fatalf("error: Expected the client to be empty: %v", client)
	} else {
		assert.Equal(t, err.Error(), "error: required environment variable 'VO_TF_API_KEY' is not set")
	}
}

func configureTestProvider(testProvider terraform.ResourceProvider, raw map[string]interface{}) (Config, error) {
	err := testProvider.Configure(terraform.NewResourceConfigRaw(raw))
	if meta := testProvider.(*schema.Provider).Meta(); meta == nil {
		return Config{}, err
	} else {
		return meta.(Config), nil
	}
}

func TestProviderConfigureFromNothing(t *testing.T) {

	raw := make(map[string]interface{})

	if configuration, err := configureTestProvider(testAccProviders["victorops_config_test"], raw); err != nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	} else {
		assert.Equal(t, "", configuration.APIId)
		assert.Equal(t, "", configuration.APIKey)
		assert.Equal(t, "https://api.victorops.com", configuration.BaseURL)
	}
}

func TestProviderConfigureFromTerraform(t *testing.T) {

	raw := map[string]interface{}{
		"api_id":  "515151",
		"api_key": "151515",
	}

	if configuration, err := configureTestProvider(testAccProviders["victorops_config_test"], raw); err != nil {
		t.Fatalf("Expected metadata, got nil. err: %s", err.Error())
	} else {
		assert.Equal(t, "515151", configuration.APIId)
		assert.Equal(t, "151515", configuration.APIKey)
		assert.Equal(t, "https://api.victorops.com", configuration.BaseURL)
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

// TODO: Set-up the client connection to test-org for initiating Acceptance tests
func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}
