package db

import (
	"strings"
	"time"

	"golang.org/x/xerrors"

	c "github.com/vulsio/goval-dictionary/config"
	"github.com/vulsio/goval-dictionary/models"
)

// DB is interface for a database driver
type DB interface {
	Name() string
	OpenDB(string, string, bool, Option) (bool, error)
	CloseDB() error
	MigrateDB() error

	IsGovalDictModelV1() (bool, error)
	GetFetchMeta() (*models.FetchMeta, error)
	UpsertFetchMeta(*models.FetchMeta) error

	GetByPackName(family string, osVer string, packName string, arch string) ([]models.Definition, error)
	GetByCveID(family string, osVer string, cveID string, arch string) ([]models.Definition, error)
	InsertOval(*models.Root) error
	CountDefs(string, string) (int, error)
	GetLastModified(string, string) (time.Time, error)
}

// Option :
type Option struct {
	RedisTimeout time.Duration
}

// NewDB return DB accessor.
func NewDB(dbType, dbPath string, debugSQL bool, option Option) (driver DB, locked bool, err error) {
	if driver, err = newDB(dbType); err != nil {
		return driver, false, xerrors.Errorf("Failed to new db. err: %w", err)
	}

	if locked, err := driver.OpenDB(dbType, dbPath, debugSQL, option); err != nil {
		if locked {
			return nil, true, err
		}
		return nil, false, err
	}

	isV1, err := driver.IsGovalDictModelV1()
	if err != nil {
		return nil, false, xerrors.Errorf("Failed to IsGovalDictModelV1. err: %w", err)
	}
	if isV1 {
		return nil, false, xerrors.New("Failed to NewDB. Since SchemaVersion is incompatible, delete Database and fetch again.")
	}

	if err := driver.MigrateDB(); err != nil {
		return driver, false, xerrors.Errorf("Failed to migrate db. err: %w", err)
	}
	return driver, false, nil
}

func newDB(dbType string) (DB, error) {
	switch dbType {
	case dialectSqlite3, dialectMysql, dialectPostgreSQL:
		return &RDBDriver{name: dbType}, nil
	case dialectRedis:
		return &RedisDriver{name: dbType}, nil
	}
	return nil, xerrors.Errorf("Invalid database dialect. dbType: %s", dbType)
}

func formatFamilyAndOSVer(family, osVer string) (string, string, error) {
	switch family {
	case c.Debian:
		osVer = major(osVer)
	case c.Ubuntu:
		osVer = major(osVer)
	case c.Raspbian:
		family = c.Debian
		osVer = major(osVer)
	case c.RedHat:
		osVer = major(osVer)
	case c.CentOS:
		family = c.RedHat
		osVer = major(osVer)
	case c.Oracle:
		osVer = major(osVer)
	case c.Amazon:
		osVer = getAmazonLinuxVer(osVer)
	case c.Alpine:
		osVer = majorDotMinor(osVer)
	case c.Fedora:
		osVer = major(osVer)
	case c.OpenSUSE:
		if osVer != "tumbleweed" {
			osVer = majorDotMinor(osVer)
		}
	case c.OpenSUSELeap, c.SUSEEnterpriseDesktop, c.SUSEEnterpriseServer:
		osVer = majorDotMinor(osVer)
	default:
		return "", "", xerrors.Errorf("Failed to detect family. err: unknown os family(%s)", family)
	}

	return family, osVer, nil
}

func major(osVer string) (majorVersion string) {
	return strings.Split(osVer, ".")[0]
}

func majorDotMinor(osVer string) (majorMinorVersion string) {
	ss := strings.Split(osVer, ".")
	if len(ss) < 3 {
		return osVer
	}
	return strings.Join(ss[:2], ".")
}

// getAmazonLinuxVer returns AmazonLinux 1, 2, 2022
func getAmazonLinuxVer(osVersion string) string {
	ss := strings.Fields(osVersion)
	if ss[0] == "2022" {
		return "2022"
	}
	if ss[0] == "2" {
		return "2"
	}
	return "1"
}

// IndexChunk has a starting point and an ending point for Chunk
type IndexChunk struct {
	From, To int
}

func chunkSlice(length int, chunkSize int) <-chan IndexChunk {
	ch := make(chan IndexChunk)

	go func() {
		defer close(ch)

		for i := 0; i < length; i += chunkSize {
			idx := IndexChunk{i, i + chunkSize}
			if length < idx.To {
				idx.To = length
			}
			ch <- idx
		}
	}()

	return ch
}

func filterByRedHatMajor(packs []models.Package, majorVer string) (filtered []models.Package) {
	for _, p := range packs {
		if strings.Contains(p.Version, ".el"+majorVer) ||
			strings.Contains(p.Version, ".module+el"+majorVer) {
			filtered = append(filtered, p)
		}
	}
	return
}
