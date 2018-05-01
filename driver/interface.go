package driver

import (
	"errors"

	"github.com/chapsuk/miga/driver/goose"
	"github.com/chapsuk/miga/driver/migrate"
)

const (
	Goose   = "goose"
	Migrate = "migrate"
)

func Available(name string) bool {
	switch name {
	case Goose:
	case Migrate:
	default:
		return false
	}
	return true
}

type Config struct {
	Name string

	Dialect          string
	Dsn              string
	Dir              string
	VersionTableName string
}

type Interface interface {
	Create(name, ext string) error
	Down() error
	DownTo(version string) error
	Redo() error
	Reset() error
	Status() error
	Up() error
	UpTo(version string) error
	Version() error
}

func New(cfg *Config) (Interface, error) {
	switch cfg.Name {
	case Goose:
		return goose.New(
			cfg.Dialect,
			cfg.Dsn,
			cfg.VersionTableName,
			cfg.Dir,
		)
	case Migrate:
		return migrate.New(
			cfg.Dialect,
			cfg.Dsn,
			cfg.VersionTableName,
			cfg.Dir,
		)
	default:
		return nil, errors.New("unsupported migrations driver")
	}
}
