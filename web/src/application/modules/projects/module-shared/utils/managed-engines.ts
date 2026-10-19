/**
 * What the screens say of a managed engine, in one table: copy that names an
 * engine or what it works with reads it from here, so a second engine is a row
 * rather than a condition in each screen.
 */
export interface ManagedEngineCopy {
    /** The engine as a person names it: "PostgreSQL". */
    name: string;
    /** What the engine archives continuously, as a sentence names it: "WAL", "binary logs". */
    archiveNoun: string;
    /** Whether archiveNoun takes a plural verb: "the binary logs are", but "the WAL is". */
    archiveNounPlural: boolean;
    /** The tool the dump runs: "pg_dump". */
    dumpTool: string;
    /** A check query a restore test might run, as a placeholder shows it. */
    checkQueryExample: string;
}

const MANAGED_ENGINES: Record<string, ManagedEngineCopy> = {
    postgres: {
        name: "PostgreSQL",
        archiveNoun: "WAL",
        archiveNounPlural: false,
        dumpTool: "pg_dump",
        checkQueryExample: "SELECT count(*) FROM orders",
    },
    mariadb: {
        name: "MariaDB",
        archiveNoun: "binary logs",
        archiveNounPlural: true,
        dumpTool: "mariadb-dump",
        checkQueryExample: "SELECT COUNT(*) FROM orders",
    },
};

/** An engine's copy. One this dashboard does not know is named as it is, and the rest said in general words. */
export function managedEngineOf(engine: string): ManagedEngineCopy {
    return (
        MANAGED_ENGINES[engine] ?? {
            name: engine,
            archiveNoun: "logs",
            archiveNounPlural: true,
            dumpTool: "logical dump",
            checkQueryExample: "SELECT 1",
        }
    );
}
