package sdk

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go/aws"
	"gopkg.in/ini.v1"
)

type KeychainConfig struct {
	API      string
	name     string
	clientID string
	Username string
	password string
	region   string
	account  string
}

func (k *Keychain) GetAccount() string {
	return k.account
}

func (k *Keychain) SetAccount(account string) {
	k.account = account
}

func NewKeychainFromIniFile(profile string) *Keychain {
	cfg, err := ini.Load(os.ExpandEnv("$HOME/.praetorian/keychain.ini"))
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	sections := cfg.Sections()

	var section *ini.Section
	if profile != "" {
		section, err = cfg.GetSection(profile)
		if err != nil {
			log.Fatalf("Failed to find profile %s in keychain.ini: %v", profile, err)
		}
	}
	if section == nil && len(sections) >= 2 {
		// section[0] is a default empty section and should be skipped
		section = sections[1]
	} else if section == nil {
		log.Fatalf("No sections found in keychain.ini")
	}

	config := KeychainConfig{
		API:      section.Key("api").String(),
		name:     section.Key("name").String(),
		clientID: section.Key("client_id").String(),
		Username: section.Key("username").String(),
		password: section.Key("password").String(),
		account:  "",
		region:   "us-east-2",
	}
	return newKeychain(config)
}

type Keychain struct {
	KeychainConfig
	tokenCache  string
	tokenExpiry int64
}

func newKeychain(config KeychainConfig) *Keychain {
	return &Keychain{
		KeychainConfig: config,
	}
}

func (k *Keychain) GetToken() (string, error) {
	if k.tokenCache == "" || time.Now().Unix() >= k.tokenExpiry {
		cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(k.region))
		if err != nil {
			return "", fmt.Errorf("unable to load SDK config, %v", err)
		}

		client := cognitoidentityprovider.NewFromConfig(cfg)

		input := &cognitoidentityprovider.InitiateAuthInput{
			AuthFlow: "USER_PASSWORD_AUTH",
			AuthParameters: map[string]string{
				"USERNAME": k.Username,
				"PASSWORD": k.password,
			},
			ClientId: aws.String(k.clientID),
		}

		resp, err := client.InitiateAuth(context.TODO(), input)
		if err != nil {
			return "", fmt.Errorf("failed to initiate auth, %v", err)
		}

		k.tokenExpiry = time.Now().Unix() + int64(resp.AuthenticationResult.ExpiresIn)
		k.tokenCache = *resp.AuthenticationResult.IdToken
	}

	return k.tokenCache, nil
}
