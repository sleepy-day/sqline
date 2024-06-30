package app

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

const (
	LINUX_CONF   = "$HOME/.config/sqline"
	MACOS_CONF   = "$HOME/Library/Application Support/sqline"
	WINDOWS_CONF = "C:\\Users\\%USER%\\AppData\\Roaming\\sqline"
)

type ConnInfo struct {
	Name    string `toml:"name"`
	Driver  string `toml:"driver"`
	ConnStr string `toml:"connstr"`
}

type Connections struct {
	Conns []ConnInfo `toml:"connections"`
}

func loadConns() *Connections {
	var conf string
	switch runtime.GOOS {
	case "linux":
		conf = LINUX_CONF
	case "darwin":
		conf = MACOS_CONF
	case "windows":
		conf = WINDOWS_CONF
	}

	if conf == "" {
		return nil
	}

	fp := filepath.Join(conf, "conns.toml")
	f, err := os.Open(fp)
	if err != nil {
		return nil
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return nil
	}

	if len(data) == 0 {
		return nil
	}

	var conns Connections
	err = toml.Unmarshal(data, &conns)
	if err != nil {
		panic("Error loading connection toml")
	}

	return &conns
}

func saveConns() error {
	var conf string
	switch runtime.GOOS {
	case "linux":
		conf = LINUX_CONF
	case "darwin":
		conf = MACOS_CONF
	case "windows":
		conf = WINDOWS_CONF
	}

	if conf == "" {
		return errors.New("Unable to identify platform")
	}

	file := filepath.Join(conf, "conns.toml")

	tomlData, err := toml.Marshal(savedConns)
	if err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(tomlData)
	return err
}
