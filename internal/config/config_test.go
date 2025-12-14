package config

import (
	"os"
	"path/filepath"
	"testing"

	"olexsmir.xyz/x/is"
)

func TestNewConfig(t *testing.T) {
}

func TestParsePassword(t *testing.T) {
	passwd := "qwerty123"

	t.Run("string", func(t *testing.T) {
		r, err := parsePassword(passwd, ".")
		is.Err(t, err, nil)
		is.Equal(t, r, passwd)
	})

	t.Run("env var", func(t *testing.T) {
		t.Setenv("secret_password", passwd)
		r, err := parsePassword("$env:secret_password", ".")
		is.Err(t, err, nil)
		is.Equal(t, r, passwd)
	})

	t.Run("unset env var", func(t *testing.T) {
		_, err := parsePassword("$env:secret_password", ".")
		is.Err(t, err, ErrUnsetPasswordEnv)
	})

	t.Run("file", func(t *testing.T) {
		r, err := parsePassword("file:./testdata/password", ".")
		is.Err(t, err, nil)
		is.Equal(t, r, passwd)
	})

	t.Run("empty file", func(t *testing.T) {
		_, err := parsePassword("file:./testdata/empty_password", ".")
		is.Err(t, err, ErrEmptyPasswordFile)
	})

	t.Run("non existing file", func(t *testing.T) {
		_, err := parsePassword("file:/not/exists", ".")
		is.Err(t, err, ErrPasswordFileNotFound)
	})

	t.Run("file, not set path", func(t *testing.T) {
		_, err := parsePassword("file:", ".")
		is.Err(t, err, ErrPasswordFileNotFound)
	})

	t.Run("file, path with env", func(t *testing.T) {
		tmpdir := t.TempDir()
		t.Setenv("TMP_DIR", tmpdir)

		data, _ := os.ReadFile("./testdata/password")
		os.WriteFile(filepath.Join(tmpdir, "password"), data, 0o644)

		r, err := parsePassword("file:$TMP_DIR/password", ".")
		is.Err(t, err, nil)
		is.Equal(t, r, passwd)
	})
}
