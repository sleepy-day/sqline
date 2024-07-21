package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"
)

type ConnInfo struct {
	Name    string `toml:"name"`
	Driver  string `toml:"driver"`
	ConnStr string `toml:"connstr"`
}

func LinuxConf() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s/.config/sqline", home)
}

func MacosConf() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s/Library/Application Support/sqline", home)
}

func WindowsConf() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%s\\AppData\\Roaming\\sqline", home)
}

type Connections struct {
	Conns []ConnInfo `toml:"connections"`
}

func (conInf ConnInfo) Text() string {
	return conInf.Name
}

func (conInf ConnInfo) Subtext() string {
	return conInf.Driver
}

//TODO: Handle no home dir found

func loadConns() *Connections {
	conns := Connections{}

	var conf string
	switch runtime.GOOS {
	case "linux":
		conf = LinuxConf()
	case "darwin":
		conf = MacosConf()
	case "windows":
		conf = WindowsConf()
	}

	if conf == "" {
		return &conns
	}

	fp := filepath.Join(conf, "conns.toml")
	f, err := os.Open(fp)
	if err != nil {
		return &conns
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return &conns
	}

	if len(data) == 0 {
		return &conns
	}

	err = toml.Unmarshal(data, &conns)
	if err != nil {
		panic("Error loading connection toml")
	}

	return &conns
}

func saveConns(conns *Connections) error {
	var conf string
	switch runtime.GOOS {
	case "linux":
		conf = LinuxConf()
	case "darwin":
		conf = MacosConf()
	case "windows":
		conf = WindowsConf()
	}

	if conf == "" {
		return errors.New("Unable to identify platform")
	}

	file := filepath.Join(conf, "conns.toml")

	tomlData, err := toml.Marshal(conns)
	if err != nil {
		return err
	}

	err = os.MkdirAll(conf, os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(tomlData)
	if err != nil {
		return err
	}

	ConnList.SetOptions(conns.Conns)

	return nil
}
