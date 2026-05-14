package storage

func NewSQLiteProvider(dsn string) (Provider, error) {
	return newSQLStore(dsn, "sqlite"), nil
}
