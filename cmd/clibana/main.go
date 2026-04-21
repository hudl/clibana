package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
)

const (
	MinPollInterval   = 0
	MaxPollInterval   = 5 * time.Second
	ResponseTimeout   = 10
	AuthTypeAWS       = "aws"
	AuthTypeBasic     = "basic"
	AuthTypeCookie    = "cookie"
	SearchRequestSize = 10000
	HitChannelBuffer  = 30000
	ServerTypeOpenSearch     = "opensearch"
	ServerTypeElasticSearch = "elasticsearch"
)

var (
	version = "1.3.0"
	commit  = ""
	date    = ""
)

var DebugLogger = log.New(io.Discard, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

func FatalError(err error) {
	fmt.Fprintf(os.Stderr, "clibana: %v\n", err)
	os.Exit(1)
}

func handleKeychainSet() {
	for i, arg := range os.Args[1:] {
		if arg == "--password-keychain-set" {
			if i+1 >= len(os.Args[1:]) {
				FatalError(fmt.Errorf("--password-keychain-set requires SERVICE:ACCOUNT argument"))
			}
			target := os.Args[i+2]
			parts := strings.SplitN(target, ":", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				FatalError(fmt.Errorf("--password-keychain-set requires format SERVICE:ACCOUNT (e.g., snowplow-kibana:hudl_admin)"))
			}
			password, err := io.ReadAll(os.Stdin)
			if err != nil {
				FatalError(fmt.Errorf("failed to read password from stdin: %w", err))
			}
			pw := strings.TrimSpace(string(password))
			if pw == "" {
				FatalError(fmt.Errorf("empty password received on stdin"))
			}
			if err := keyring.Set(parts[0], parts[1], pw); err != nil {
				FatalError(fmt.Errorf("failed to store password in keychain (service=%q, account=%q): %w", parts[0], parts[1], err))
			}
			fmt.Fprintln(os.Stderr, "Password stored in keychain")
			os.Exit(0)
		}
	}
}

func main() {
	handleKeychainSet()

	clibanaConfig := NewClibanaConfig()

	if clibanaConfig.Debug {
		DebugLogger.SetOutput(os.Stderr)

	}

	client, err := createClient(clibanaConfig)
	if err != nil {
		FatalError(fmt.Errorf("Failed to create OpenSearch client: %w", err))
	}

	DebugLogger.Printf("Configuration: %+v\n", clibanaConfig)

	switch {
	case clibanaConfig.Search != nil:
		search(client, clibanaConfig)
	case clibanaConfig.Mappings != nil:
		mappings(client, clibanaConfig)
	case clibanaConfig.Indices != nil:
		indices(client, clibanaConfig)
	default:
		FatalError(fmt.Errorf("no subcommand specified"))
	}

}
