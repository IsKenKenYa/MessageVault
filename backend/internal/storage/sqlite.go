package storage

func NewSQLiteProvider(dsn string) (Provider, error) {
	return newFileStore(dsn, "filestore/sqlite-adapter"), nil
}
