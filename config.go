package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Config struct {
	ZFSPool       string `json:"zfs_pool"`
	ZFSBackupPool string `json:"zfs_backup_pool"`
}

func ReadConfig() Config {
	var config Config

	configDir, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	content, err := ioutil.ReadFile(filepath.Join(configDir, "zfsbackup", "config.json"))
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(content, &config)
	if err != nil {
		panic(err)
	}

	return config
}

func (c *Config) WriteConfig() {
	content, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("~/.config/zfsbackup/config.json", content, 0644)
	if err != nil {
		panic(err)
	}
}
