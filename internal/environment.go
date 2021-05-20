package internal

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type (
	Environ struct {
		ServerPort int           `required:"true" envconfig:"PORT"`
		SkipChecks bool          `required:"true" envconfig:"SKIP_CHECKS"`
		Secure     bool          `required:"true" envconfig:"SECURE"`
		JwtSecret  string        `required:"true" envconfig:"SECRET_KEY"`
		JwtExp     time.Duration `required:"true" envconfig:"JWT_EXP"`
		DBUri      string        `required:"true" envconfig:"DB_URI"`
		DBName     string        `required:"true" envconfig:"DB_NAME"`
		DBRootUser string        `required:"true" envconfig:"DB_ROOT_USER"`
		DBRootPass string        `required:"true" envconfig:"DB_ROOT_PASS"`
		SMTPHost   string        `required:"false" envconfig:"SMTP_HOST"`
		SMTPPort   int           `required:"false" envconfig:"SMTP_PORT"`
		SMTPUser   string        `required:"false" envconfig:"SMTP_USER"`
		SMTPPass   string        `required:"false" envconfig:"SMTP_PASS"`
		FEndpoint  string        `required:"true" envconfig:"FE_ENDPOINT"`
	}
)

// GetEnv gets required environment variables
func GetEnv() (*Environ, error) {
	e := new(Environ)

	// Process and validate the required environment variables
	err := envconfig.Process("INST", e)
	if err != nil {
		return nil, err
	}

	return e, nil
}
