package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	BoolOptions   map[string]bool
	StringOptions map[string]string
	IntOptions    map[string]int
}

func parseConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{
		BoolOptions:   make(map[string]bool),
		StringOptions: make(map[string]string),
		IntOptions:    make(map[string]int),
	}

	scanner := bufio.NewScanner(file)
	var currentSection string
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		split := strings.SplitN(line, "=", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}

		key := split[0]
		value := split[1]

		switch currentSection {
		case "bool":
			boolValue, err := strconv.ParseBool(value)
			if err != nil {
				return nil, fmt.Errorf("invalid bool value for key %s: %v", key, err)
			}
			config.BoolOptions[key] = boolValue
		case "string":
			config.StringOptions[key] = value
		case "int":
			intValue, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid int value for key %s: %v", key, err)
			}
			config.IntOptions[key] = intValue
		default:
			return nil, fmt.Errorf("unknown section: %s", currentSection)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return config, nil
}
