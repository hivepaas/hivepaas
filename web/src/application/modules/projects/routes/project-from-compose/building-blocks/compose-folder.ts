/** The names a compose file goes by, in the order compose looks for them. */
const COMPOSE_NAMES = ["compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"];

/** The names an example .env goes by beside a compose file. */
const ENV_EXAMPLE_NAMES = [".env.example", "example.env", ".env.sample", "env.example"];

/** Files a folder holds for git or the desktop: never what a directory gives an app. */
const JUNK_NAMES = new Set([".gitkeep", ".gitignore", ".keep", ".DS_Store", "Thumbs.db", "desktop.ini"]);

/** The largest file the server takes, as a config's: Docker's limit. */
export const COMPOSE_FILE_MAX_BYTES = 500 * 1024;

/** How many files, and how many bytes of them, a request carries at most. */
export const COMPOSE_FILES_MAX = 100;
export const COMPOSE_FILES_MAX_BYTES = 5 * 1024 * 1024;

/**
 * A folder opened in the browser: its compose file, and the files beside it.
 * Nothing of it leaves the browser but the files the compose file reads, as
 * the review asks for them.
 */
export interface ComposeFolder {
    /** The folder's own name. */
    name: string;
    /** The compose file's directory in the folder; empty for the folder itself. */
    base: string;
    compose: File;
    dotEnv?: File;
    /** An example .env beside the compose file, offered when there is no .env. */
    envExample?: File;
    /** Every file under the compose file's directory, by its path relative to it. */
    files: Map<string, File>;
}

/** The path of a file in the folder the browser gave, without the folder's own name. */
function pathInFolder(file: File): { folder: string; path: string } {
    const relative = file.webkitRelativePath || file.name;
    const slash = relative.indexOf("/");

    return slash < 0
        ? { folder: "", path: relative }
        : { folder: relative.slice(0, slash), path: relative.slice(slash + 1) };
}

function dirOf(path: string): string {
    const slash = path.lastIndexOf("/");

    return slash < 0 ? "" : path.slice(0, slash);
}

function baseName(path: string): string {
    return path.slice(path.lastIndexOf("/") + 1);
}

/**
 * Reads a folder as compose would: the compose file is the shallowest one, by
 * compose's own order of names, and every other path is relative to its
 * directory. undefined when the folder has none.
 */
export function readComposeFolder(list: FileList | File[]): ComposeFolder | undefined {
    const entries = Array.from(list).map(file => ({ file, ...pathInFolder(file) }));
    let best: { file: File; path: string; depth: number; rank: number } | undefined;
    for (const entry of entries) {
        const rank = COMPOSE_NAMES.indexOf(baseName(entry.path));
        if (rank < 0) {
            continue;
        }
        const dir = dirOf(entry.path);
        const depth = dir === "" ? 0 : dir.split("/").length;
        if (!best || depth < best.depth || (depth === best.depth && rank < best.rank)) {
            best = { file: entry.file, path: entry.path, depth, rank };
        }
    }
    if (!best) {
        return undefined;
    }

    const base = dirOf(best.path);
    const prefix = base === "" ? "" : `${base}/`;
    const files = new Map<string, File>();
    for (const entry of entries) {
        if (entry.path !== best.path && entry.path.startsWith(prefix)) {
            files.set(entry.path.slice(prefix.length), entry.file);
        }
    }
    const envExample = ENV_EXAMPLE_NAMES.map(name => files.get(name)).find(file => file !== undefined);

    return {
        name: entries[0]?.folder ?? "",
        base,
        compose: best.file,
        dotEnv: files.get(".env"),
        envExample,
        files,
    };
}

/**
 * The files of the folder the review asks for that are not given yet, within
 * what a request carries: none larger than a config may be, and no more, nor
 * more bytes, than the server takes with those already given.
 */
export function folderFilesFor(
    folder: ComposeFolder,
    paths: string[],
    given: Record<string, File | string>,
    skip: Set<string>,
): Record<string, File> {
    let count = Object.keys(given).length;
    let bytes = Object.values(given).reduce(
        (total, file) => total + (typeof file === "string" ? file.length : file.size),
        0,
    );
    const out: Record<string, File> = {};
    for (const path of paths) {
        const file = folder.files.get(path);
        if (!file || path in given || skip.has(path) || file.size > COMPOSE_FILE_MAX_BYTES) {
            continue;
        }
        if (count + 1 > COMPOSE_FILES_MAX || bytes + file.size > COMPOSE_FILES_MAX_BYTES) {
            break;
        }
        out[path] = file;
        count += 1;
        bytes += file.size;
    }

    return out;
}

/** The files of the folder under one of its directories, by path, but those for git or the desktop. */
export function folderFilesUnder(folder: ComposeFolder, dir: string): [string, File][] {
    const prefix = `${dir}/`;

    return [...folder.files.entries()]
        .filter(([path]) => path.startsWith(prefix) && !JUNK_NAMES.has(baseName(path)))
        .sort(([a], [b]) => a.localeCompare(b));
}

/**
 * The files of the folder under a directory the review lists, to give: all of
 * them or none - a directory half given is not what the folder has - within
 * what a request carries with those already given.
 */
export function folderDirectoryFiles(
    folder: ComposeFolder,
    dir: string,
    given: Record<string, File | string>,
): { files: Record<string, File>; fits: boolean } {
    const under = folderFilesUnder(folder, dir).filter(([path]) => !(path in given));
    const count = Object.keys(given).length + under.length;
    const bytes =
        Object.values(given).reduce((total, file) => total + (typeof file === "string" ? file.length : file.size), 0) +
        under.reduce((total, [, file]) => total + file.size, 0);
    const fits =
        count <= COMPOSE_FILES_MAX &&
        bytes <= COMPOSE_FILES_MAX_BYTES &&
        under.every(([, file]) => file.size <= COMPOSE_FILE_MAX_BYTES);

    return { files: fits ? Object.fromEntries(under) : {}, fits };
}
