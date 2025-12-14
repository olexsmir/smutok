package config

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
)

//go:embed config.toml
var defaultConfig []byte

var (
	ErrUnsetPasswordEnv     = errors.New("password env is unset")
	ErrNotInitializedConfig = errors.New("config is not initialized")
	ErrConfigAlreadyExists  = errors.New("config already exists")
	ErrPasswordFileNotFound = errors.New("password file not found")
	ErrEmptyPasswordFile    = errors.New("password file is empty")
)

type Config struct {
	DBPath      string
	LogFilePath string
	FreshRSS    struct {
		Host     string `toml:"host"`
		Username string `toml:"username"`
		Password string `toml:"password"`
	} `toml:"freshrss"`
}

func New() (*Config, error) {
	configPath := MustGetConfigFilePath()
	if !isFileExists(configPath) {
		return nil, ErrNotInitializedConfig
	}

	configRaw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config *Config
	if cerr := toml.Unmarshal(configRaw, &config); cerr != nil {
		return nil, cerr
	}

	passwd, err := parsePassword(
		config.FreshRSS.Password,
		filepath.Dir(configPath))
	if err != nil {
		return nil, err
	}

	config.FreshRSS.Password = passwd
	config.DBPath = mustGetStateFile("smutok.sqlite")
	config.LogFilePath = mustGetStateFile("smutok.log")

	return config, nil
}

func Init() error {
	configPath := MustGetConfigFilePath()
	if isFileExists(configPath) {
		return ErrConfigAlreadyExists
	}

	err := os.WriteFile(configPath, defaultConfig, 0o644)
	return err
}

func MustGetConfigFilePath() string { return mustGetConfigFile("config.toml") }

func mustGetStateFile(file string) string {
	stateFile, err := xdg.StateFile("smutok/" + file)
	if err != nil {
		panic(err)
	}
	return stateFile
}

func mustGetConfigFile(file string) string {
	configFile, err := xdg.ConfigFile("smutok/" + file)
	if err != nil {
		panic(err)
	}
	return configFile
}

func parsePassword(passwd string, baseDir string) (string, error) {
	envPrefix := "$env:"
	filePrefix := "file:"

	switch {
	case strings.HasPrefix(passwd, envPrefix):
		env := os.Getenv(passwd[len(envPrefix):])
		if env == "" {
			return "", ErrUnsetPasswordEnv
		}
		return env, nil

	case strings.HasPrefix(passwd, filePrefix):
		fpath := os.ExpandEnv(passwd[len(filePrefix):])

		if strings.HasPrefix(fpath, "./") {
			fpath = filepath.Join(baseDir, fpath)
		}

		if !isFileExists(fpath) {
			return "", ErrPasswordFileNotFound
		}

		data, err := os.ReadFile(fpath)
		if err != nil {
			return "", err
		}

		password := strings.TrimSpace(string(data))
		if password == "" {
			return "", ErrEmptyPasswordFile
		}
		return password, nil

	default:
		return passwd, nil
	}
}

func isFileExists(fpath string) bool {
	_, err := os.Stat(fpath)
	return err == nil
}
