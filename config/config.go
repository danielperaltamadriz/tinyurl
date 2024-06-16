package config

type API struct {
	Port int
}
type DB struct {
	ConnectionString string
}
type Config struct {
	API API
	DB  DB
}
