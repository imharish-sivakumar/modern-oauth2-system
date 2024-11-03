package constants

// Environment is an enum for type Environment.
type Environment string

const (
	Local Environment = "LOCAL"
	Dev   Environment = "DEV"
	Prod  Environment = "PROD"
)

func (e Environment) IsLocal() bool {
	return e == Local
}
