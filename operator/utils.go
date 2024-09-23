package operator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func loadConfig(filename string) (*Config, error) {
	// Open the config file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	// Read the file contents
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	// Unmarshal the JSON data into the Config struct
	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}
