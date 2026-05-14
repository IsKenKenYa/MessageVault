package storage

func NewPostgresProvider(dsn string) (Provider, error) {
	return newSQLStore(dsn, "postgres"), nil
}
