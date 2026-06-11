//nolint:all
package config

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func readerFunc(path string) ([]byte, error) {
	if path == "/correct/path/config.test.yml" {
		return []byte("work_mode: webapp\npostgres:\n  host: postgres\n  user: postgres\n  password: postgres\n  dbname: app\n  port: 5432"), nil
	} else if path == "/incorrect/path" {
		return nil, errors.New("open ./incorrect/path/config.test.yml: no such file or directory")
	}
	return nil, errors.New("path is unpredictable")
}

func TestConfig(t *testing.T) {
	env := "test"
	cnf := NewConfig(env)
	assert.Equal(t, cnf.Environment, env)
	err := cnf.ReadConfig("/correct/path", readerFunc)
	assert.Nil(t, err)
	err = cnf.ReadConfig("/incorrect/path", readerFunc)
	assert.NotNil(t, err)
}
