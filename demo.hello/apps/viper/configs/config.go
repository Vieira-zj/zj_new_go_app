package configs

// Configurations exported
type Configurations struct {
	Server      ServerConfigurations
	Database    DatabaseConfigurations
	ExamplePath string `yaml:"EXAMPLE_PATH"`
	ExampleVar  string `yaml:"EXAMPLE_VAR"`
}

// ServerConfigurations exported
type ServerConfigurations struct {
	Port int
}

// DatabaseConfigurations exported
type DatabaseConfigurations struct {
	DBName     string
	DBUser     string
	DBPassword string
}
