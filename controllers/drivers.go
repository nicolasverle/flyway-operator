package controllers

import (
	"fmt"

	migrationsv1alpha1 "flyway-operator/api/v1alpha1"

	"github.com/jmoiron/sqlx"

	// import postgres driver
	_ "github.com/lib/pq"
)

type (
	// Driver interface
	Driver interface {
		CheckDBAvailability(spec *migrationsv1alpha1.DBSpec, creds *UserPassword) (bool, error)
		ConnectionURL(spec *migrationsv1alpha1.DBSpec) string
	}

	// PostgresDriver implementation
	PostgresDriver struct{}
)

var (
	Drivers = map[string]Driver{
		"org.postgresql.Driver": PostgresDriver{},
	}
)

func (d PostgresDriver) CheckDBAvailability(spec *migrationsv1alpha1.DBSpec, creds *UserPassword) (bool, error) {
	_, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s", spec.Host, spec.Port, spec.DBName, creds.User, creds.Password))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d PostgresDriver) ConnectionURL(spec *migrationsv1alpha1.DBSpec) string {
	return fmt.Sprintf("jdbc:postgresql://%s:%d/%s", spec.Host, spec.Port, spec.DBName)
}
