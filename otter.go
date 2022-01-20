package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joshuamulliken/transcribe/pkg/otterapi"
	"github.com/shibukawa/configdir"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type credSource int

const (
	// credSourceNone means that no credentials were provided
	credSourceNone credSource = iota
	// credSourceEnv means that credentials were provided via environment variables
	credSourceEnv
	// credSourceFile means that credentials were provided via a config file (see -c)
	credSourceFile
	// credSourceArg means that credentials were provided via command line arguments
	credSourceArg
	// credSourcePrompt means that credentials were provided via interactive prompt
	credSourcePrompt
)

const defaultConfigFileName = "settings.json"

// otterConfig defines the configuration for the transcribe CLI
// It holds the information required to connect to the service and holds a reference to where
// it should be stored
type otterConfig struct {
	// Path to the config file in the filesystem
	Path string
	// CredSource is where the application was provided credentials
	CredSource credSource
	// Username is the username to otter.ai
	Username string `json:"username"`
	// Password is the password to otter.ai
	Password string `json:"password"`
	// SessionID is the current session ID used to authenticate api calls
	SessionID string `json:"session_id"`
}

// defaultConfigDir returns the default configuration directory based on the users OS.
func defaultConfigDir() string {
	configDirectory := configdir.New("mulliken.net", "transcribe")
	folder := configDirectory.QueryFolders(configdir.Global)[0].Path

	return folder
}

// defaultConfigFile returns the default configuration file path
func defaultConfigFile() string {
	return defaultConfigDir() + "/" + defaultConfigFileName
}

// getConfigFromPath looks for a config file at a given path and if it exists,
// parses and returns that config. Otherwise, it will look for a config at the
// default location (see defaultConfigFile) and return that.
func getConfigFromPath(customPath *string) (otterConfig, error) {
	var config otterConfig

	if *customPath != "" {
		config, err := parseConfig(customPath)
		if err != nil {
			return config, err
		}

		return config, nil
	}
	defaultConfigPath := defaultConfigFile()
	config, err := parseConfig(&defaultConfigPath)
	if err != nil {
		return otterConfig{}, err
	}

	return config, nil
}

func parseConfig(configPath *string) (otterConfig, error) {
	configFile, err := os.Open(*configPath)
	if err != nil {
		return otterConfig{}, err
	}
	defer func(configFile *os.File) {
		err := configFile.Close()
		if err != nil {
			panic(err)
		}
	}(configFile)

	var config otterConfig

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		return otterConfig{}, err
	}

	config.Path = *configPath
	config.CredSource = credSourceFile

	return config, nil
}

func writeConfig(config otterConfig) error {
	configBuffer := &bytes.Buffer{}
	jsonEncoder := json.NewEncoder(configBuffer)
	err := jsonEncoder.Encode(config)
	if err != nil {
		return err
	}

	if config.Path == "" {
		config.Path = defaultConfigFile()
	}

	// Create the directory if it doesn't exist
	err = os.MkdirAll(filepath.Dir(config.Path), 0700)
	if err != nil {
		return err
	}

	// Write the config file
	err = os.WriteFile(config.Path, configBuffer.Bytes(), 0600)
	if err != nil {
		return err
	}

	return nil
}

func getConfig(argUsername *string, argPassword *string, argConfigPath *string) (otterConfig, error) {
	// If we got credentials from the command line, use those
	if *argUsername != "" && *argPassword != "" {
		config := otterConfig{
			Username:   *argUsername,
			Password:   *argPassword,
			CredSource: credSourceArg,
			Path:       defaultConfigFile(),
		}
		return config, nil
	}

	// If we got credentials from the environment, use those
	envUsername, unameOk := os.LookupEnv("OTTER_USERNAME")
	envPassword, passOk := os.LookupEnv("OTTER_PASSWORD")
	if (unameOk && passOk) && (envUsername != "" && envPassword != "") {
		config := otterConfig{
			Username:   envUsername,
			Password:   envPassword,
			CredSource: credSourceEnv,
			Path:       defaultConfigFile(),
		}
		return config, nil
	}

	// If we were given a config file, use that
	if *argConfigPath != "" {
		config, err := getConfigFromPath(argConfigPath)
		if err != nil {
			return otterConfig{}, err
		}

		return config, nil
	}

	// Otherwise, check if there is a config file at the default location
	config, err := getConfigFromPath(argConfigPath)
	if err != nil {
		// There was no config file, so get the config from the user

		// Prompt the user for credentials
		fmt.Println("Enter your otterapi.ai credentials:")
		fmt.Print("Username: ")
		username, err := readInput()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Print("Password: ")
		password, err := readInput()
		if err != nil {
			log.Fatal(err)
		}

		config = otterConfig{
			Username:   username,
			Password:   password,
			CredSource: credSourcePrompt,
			Path:       defaultConfigFile(),
		}

		return config, nil
	}

	return config, nil
}

func main() {
	// Create a command line parser
	parser := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	// Configure the flag options
	var username string
	parser.StringVar(&username, "u", "", "Username for otterapi.ai.")
	var password string
	parser.StringVar(&password, "p", "", "Password for otterapi.ai.")
	var configPath string
	parser.StringVar(&configPath, "c", "",
		fmt.Sprintf("Path to a custom config file.\nIf not specified, defaults to: %s", defaultConfigFile()))
	var shouldWriteConfig bool
	parser.BoolVar(&shouldWriteConfig, "w", false, "Write a config file to the default location.")

	parser.Usage = func() {
		fmt.Println("Usage:")
		fmt.Printf("  %s [options] <file>\n", os.Args[0])
		fmt.Println("")
		fmt.Println("Options:")
		parser.PrintDefaults()
	}

	if err := parser.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	config, err := getConfig(&username, &password, &configPath)
	if err != nil {
		log.Println("Error getting config")
		panic(err)
	}

	if config.Username == "" || config.Password == "" {
		log.Println("Credentials provided are blank.")
		log.Printf("edit \"%s\" or see -h for usage.", config.Path)
		os.Exit(1)
	}

	sessionID, err := otterapi.Login(&config.Username, &config.Password)
	if err != nil {
		log.Println("Unable to login with provided credentials")
		panic(err)
	}

	if shouldWriteConfig {
		if err := writeConfig(config); err != nil {
			log.Println("Unable to write config file")
			panic(err)
		}
	}

	if parser.NArg() != 1 {
		fmt.Println("No file provided. See -h for usage.")
		os.Exit(1)
	}

	filePath := os.Args[len(os.Args)-1]
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("Unable to open file")
		panic(err)
	}

	log.Printf("Found file: %s\n", file.Name())
	log.Println("Uploading...")
	transcriptUrl, err := otterapi.UploadSpeech(sessionID, file)
	if err != nil {
		panic(err)
	}

	log.Println("Transcript URL:")
	fmt.Println(transcriptUrl)
}

func readInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(text), nil
}
