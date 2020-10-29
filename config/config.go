package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

func NewServices() *Services {
	var services Services
	services.Openfx.FxGatewayURL = DefaultGatewayURL

	services.Functions = make(map[string]Function, 0)

	return &services
}

func ParseConfigFile(file string) (*Services, error) {
	fileData, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var services Services
	err = yaml.Unmarshal(fileData, &services)
	if err != nil {
		fmt.Printf("Error with YAML Config file\n")
		return nil, err
	}

	return &services, nil
}

func GetFxGatewayURL(argumnetURL, configURL string) string {
	var url string

	envURL := os.Getenv(GatewayEnvVarKey)

	if len(argumnetURL) > 0 && argumnetURL != DefaultGatewayURL {
		url = argumnetURL
	} else if len(configURL) > 0 && configURL != DefaultGatewayURL {
		url = configURL
	} else if len(envURL) > 0 {
		url = envURL
	} else {
		url = DefaultGatewayURL
	}

	url = strings.TrimRight(url, "/")

	return url
}

// token 발급 및 클라이언트 ID & PASSWD 파일 생성을 위해 파일 생성 및 저장 입출력 로직 필요

// New initializes a config file for the given file path
func New(filePath string) (*ConfigFile, error) {
	if filePath == "" {
		return nil, fmt.Errorf("can't create config with empty filePath")
	}
	conf := &ConfigFile{
		AuthConfigs: make([]AuthConfig, 0),
		FilePath:    filePath,
	}

	return conf, nil
}

// EnsureFile creates the root dir and config file
func EnsureFile() (string, error) {
	dirPath, err := homedir.Expand(DefaultDir)
	if err != nil {
		return "", err
	}

	filePath := path.Clean(filepath.Join(dirPath, DefaultFile))
	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
		return "", err
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			return "", err
		}
		defer file.Close()
	}
	return filePath, nil
}

// Save writes the config to disk
func (configFile *ConfigFile) save() error {
	file, err := os.OpenFile(configFile.FilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := yaml.Marshal(configFile)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return err
}

// Load reads the yml file from disk
func (configFile *ConfigFile) load() error {
	conf := &ConfigFile{}

	if _, err := os.Stat(configFile.FilePath); os.IsNotExist(err) {
		return fmt.Errorf("can't load config from non existent filePath")
	}

	data, err := ioutil.ReadFile(configFile.FilePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, conf); err != nil {
		return err
	}

	if len(conf.AuthConfigs) > 0 {
		configFile.AuthConfigs = conf.AuthConfigs
	}
	return nil
}

// UpdateAuthConfig creates or updates the username and password for a given gateway
func UpdateAuthConfig(clientID, clientSecret, token string) error {
	configPath, err := EnsureFile()
	if err != nil {
		return err
	}

	cfg, err := New(configPath) // 파일 생성 이고
	if err != nil {
		return err
	}

	if err := cfg.load(); err != nil {
		return err
	}

	auth := AuthConfig{
		Client_id:     clientID,
		Client_secret: clientSecret,
		Token:         token,
	}

	cfg.AuthConfigs[0] = auth

	if err := cfg.save(); err != nil {
		return err
	}

	return nil
}

// FileExists returns true if the config file is located at the default path
func fileExists() bool {
	dirPath, err := homedir.Expand(DefaultDir)
	if err != nil {
		return false
	}

	filePath := path.Clean(filepath.Join(dirPath, DefaultFile))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

// LookupAuthConfig returns the username and password for a given gateway
func LookupAuthConfig() (clientID, clientSecret, token string, err error) { // string, string 으로 받아야지

	if !fileExists() {
		return "", "", "", fmt.Errorf("config file does not exist.")
	}

	configPath, err := EnsureFile()
	if err != nil {
		return "", "", "", fmt.Errorf("The file path is not correct.")
	}

	cfg, err := New(configPath)
	if err != nil {
		return "", "", "", fmt.Errorf("Failed to create config data.")
	}

	if err := cfg.load(); err != nil {
		return "", "", "", fmt.Errorf("Failed to read config data.")
	}

	client_id := cfg.AuthConfigs[0].Client_id
	client_secret := cfg.AuthConfigs[0].Client_secret
	Token := cfg.AuthConfigs[0].Token

	return client_id, client_secret, Token, nil
}
