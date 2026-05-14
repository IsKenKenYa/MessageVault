package storage

func NewPostgresProvider(dsn string) (Provider, error) {
	return newFileStore(dsn, "filestore/postgres-adapter"), nil
}
