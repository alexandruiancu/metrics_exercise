package common

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ReadConfig(cfgFilePath string) map[string]string {
	config := make(map[string]string)
	file, err := os.Open(cfgFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			config[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	if !filepath.IsAbs(config["in_dir"]) ||
		!filepath.IsAbs(config["history_dir"]) {
		config["configDir"] = filepath.Dir(cfgFilePath)
	}
	return config
}
