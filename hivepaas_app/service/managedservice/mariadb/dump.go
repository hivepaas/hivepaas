package mariadb

const (
	// dumpCommand writes the app's database to stdout as SQL, consistent as of
	// its start (InnoDB), with its routines, events and triggers: what the data
	// backup streams into its repository. It reaches the server as the scripts
	// do, mysql@localhost on the socket; a shell reads the image's variables.
	dumpCommand = `sh -c 'exec gosu mysql mariadb-dump --user=mysql --single-transaction --routines --events ` +
		`--triggers --hex-blob "$MARIADB_DATABASE"'`
	// dumpRestoreCommand loads a dump from stdin into the database of the app it
	// runs in - this one, or another managed MariaDB - replacing the tables the
	// dump has. Unlike Postgres's, it is not one transaction: MariaDB commits
	// each table's DDL as it goes.
	dumpRestoreCommand = `sh -c 'exec gosu mysql mariadb --user=mysql "$MARIADB_DATABASE"'`
)
