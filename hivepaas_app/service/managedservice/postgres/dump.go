package postgres

const (
	// dumpCommand writes the database to stdout in pg_dump's custom format, as
	// the image's user over the socket: what the data backup streams into its
	// repository. A shell reads the image's variables.
	dumpCommand = `sh -c 'exec pg_dump --format=custom --username="${POSTGRES_USER:-postgres}" ` +
		`--dbname="${POSTGRES_DB:-${POSTGRES_USER:-postgres}}"'`
	// dumpRestoreCommand loads a dump from stdin into the database of the app it
	// runs in - this one, or any app of the postgres image - replacing what the
	// dump has, in one transaction: a restore that fails leaves the database as
	// it was.
	dumpRestoreCommand = `sh -c 'exec pg_restore --clean --if-exists --no-owner --single-transaction ` +
		`--username="${POSTGRES_USER:-postgres}" --dbname="${POSTGRES_DB:-${POSTGRES_USER:-postgres}}"'`
)
