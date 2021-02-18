package v1alpha1

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	// import postgres driver
	_ "github.com/lib/pq"
)

type (
	// Driver interface
	Driver interface {
		CheckDBAvailability(spec *DBSpec) (bool, error)
		ConnectionURL(spec *DBSpec) string
	}

	// PostgresDriver implementation
	PostgresDriver struct{}
)

var (
	Drivers = map[string]Driver{
		"org.postgresql.Driver": PostgresDriver{},
	}
)

func (d PostgresDriver) CheckDBAvailability(spec *DBSpec) (bool, error) {
	_, err := sqlx.Connect("postgres", fmt.Sprintf("%s:%s@%s", spec.User, spec.Password, spec.URL))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d PostgresDriver) ConnectionURL(spec *DBSpec) string {
	return fmt.Sprintf("jdbc:postgresql://%s", spec.URL)
}
