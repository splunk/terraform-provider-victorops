package victorops

import (
	"log"
	"testing"

	"github.com/victorops/go-victorops/victorops"
)

// Test config with an empty token
func TestConfigEmptyToken(t *testing.T) {
	config := Config{}

	if nullVOClient := config.VictorOpsClient; nullVOClient != nil {
		t.Fatalf("error: Expected the client to be empty: %v", nullVOClient)
	}

}

// Test config with some token
func TestConfigSkipCredsValidation(t *testing.T) {

	victoropsClient := victorops.NewClient("1234", "5678", "test_url")
	config := Config{
		VictorOpsClient: victoropsClient,
	}

	if testConfigClient := config.VictorOpsClient; testConfigClient == nil {
		t.Fatalf("error: got an empty client: %v", testConfigClient)
	} else {
		log.Printf("Client in config instantiated: %s", testConfigClient)
	}
}
