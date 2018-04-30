package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/chapsuk/miga/logger"
	"github.com/chapsuk/miga/utils"
	orig "github.com/mattes/migrate"
	"github.com/mattes/migrate/database"
	"github.com/mattes/migrate/database/mysql"
	"github.com/mattes/migrate/database/postgres"

	_ "github.com/mattes/migrate/source/file"
)

type Migrator struct {
	backend *orig.Migrate
	dir     string
	db      *sql.DB
}

func New(dialect, dsn, tableName, dir string) (*Migrator, error) {
	db, err := sql.Open(dialect, dsn)
	if err != nil {
		return nil, err
	}

	var driver database.Driver
	switch dialect {
	case "postgres":
		driver, err = postgres.WithInstance(db, &postgres.Config{
			MigrationsTable: tableName,
		})
	case "mysql":
		driver, err = mysql.WithInstance(db, &mysql.Config{
			MigrationsTable: tableName,
		})
	default:
		return nil, errors.New("Unsupported dialect")
	}

	if err != nil {
		return nil, err
	}

	m, err := orig.NewWithDatabaseInstance("file://"+dir, dialect, driver)
	if err != nil {
		return nil, err
	}
	m.Log = &utils.StdLogger{}

	return &Migrator{
		backend: m,
		dir:     dir,
		db:      db,
	}, nil
}

func (m *Migrator) Create(name, ext string) error {
	return utils.CreateMigrationsFiles(m.dir, name, ext)
}

func (m *Migrator) Down() error {
	return m.backend.Down()
}

func (m *Migrator) DownTo(version string) error {
	v, err := versionToUint(version)
	if err != nil {
		return err
	}

	current, _, err := m.backend.Version()
	if v >= current {
		return fmt.Errorf("Nothing to update, current version: %d", current)
	}

	return m.backend.Migrate(uint(v))
}

func (m *Migrator) Redo() error {
	current, dirty, err := m.backend.Version()
	if dirty {
		err = m.backend.Force(int(current))
		if err != nil {
			return err
		}
	}

	err = m.backend.Down()
	if err != nil {
		return err
	}

	return m.backend.Up()
}

func (m *Migrator) Reset() error {
	for {
		err := m.backend.Down()
		if err != nil {
			return err
		}
	}
}

func (m *Migrator) Status() error {
	return m.Version()
}

func (m *Migrator) Up() error {
	return m.backend.Up()
}

func (m *Migrator) UpTo(version string) error {
	v, err := versionToUint(version)
	if err != nil {
		return err
	}

	current, _, err := m.backend.Version()
	if v <= current {
		return fmt.Errorf("Nothing to update, current version: %d", current)
	}

	return m.backend.Migrate(uint(v))
}

func (m *Migrator) Version() error {
	version, dirty, err := m.backend.Version()
	if err != nil {
		return err
	}
	logger.G().Infof("Current version: %d dirty: %t", version, dirty)
	return nil
}

func versionToUint(version string) (uint, error) {
	v, err := strconv.Atoi(version)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}
