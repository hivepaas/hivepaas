import type { SpecImportPlan, SpecImportResult, SpecImportSelection } from "./spec-import.entity";

/**
 * A project created from a Docker Compose file: the file read into a spec
 * bundle on the server, planned and applied as an import is.
 */

/** What a published port becomes: a domain, a port on the nodes, or nothing. */
export type ComposePortAs = "domain" | "node" | "none";

export interface ComposePortView {
    published: number;
    target: number;
    protocol: string;
    /** Empty for a published port; labels for the one the service's Traefik labels route to. */
    source: string;
    as: ComposePortAs;
    /** What it becomes unless the review says otherwise. */
    default: ComposePortAs;
    domain: string;
    /** The domain offered under the root domain - or the Traefik labels' host; empty without one. */
    suggested: string;
    /** The Traefik labels' other hosts: domains of the app too, while the port is one. */
    also: string[];
}

export type ComposeVolumeKind = "volume" | "shared" | "file" | "files" | "host" | "tmpfs" | "dropped";

export interface ComposeVolumeView {
    target: string;
    /** What the file mounts: a volume's name, a path. */
    source: string;
    kind: ComposeVolumeKind;
    readOnly: boolean;
    /** The app whose directory a shared volume is. */
    owner: string;
    /** How many files given under a directory of the compose file's are mounted in it, read only. */
    files: number;
}

export interface ComposeServiceView {
    name: string;
    /** The key of the app it becomes. */
    app: string;
    image: string;
    /** The file builds its image. */
    build: boolean;
    skipped: boolean;
    reason: string;
    mode: string;
    replicas: number;
    ports: ComposePortView[];
    volumes: ComposeVolumeView[];
    /** The names it is reached by beside its key. */
    aliases: string[];
    /** Its variables written out in the file whose values are kept as its secrets. */
    secrets: string[];
    /** The fields of the file not carried to the app. */
    dropped: string[];
    /** Where the service mounts the Docker socket, which it is not given: empty for nowhere. */
    dockerSocket: string;
    /** The existing env's app its name or key is; empty for none. */
    existing: string;
    /** That app is used rather than one created. */
    useExisting: boolean;
}

export interface ComposeVariableView {
    name: string;
    default: string;
    required: boolean;
    given: boolean;
    secret: boolean;
}

export interface ComposeFileNeed {
    path: string;
    /** What reads it: env_file, config, secret, bind, directory, compose. */
    as: string;
    by: string[];
    given: boolean;
}

export interface ComposeProject {
    /** An existing project's; empty for one created. */
    id: string;
    name: string;
    key: string;
    env: string;
    envKey: string;
    /** The env is created. */
    newEnv: boolean;
    /** The project's name in the file; empty for none. */
    fileName: string;
}

export interface ComposeImportReview {
    project: ComposeProject;
    services: ComposeServiceView[];
    variables: ComposeVariableView[];
    needs: ComposeFileNeed[];
    profiles: string[];
    /** None while a required variable has no value. */
    plan?: SpecImportPlan;
}

/** A service as the app it became. */
export interface ComposeImportApp {
    service: string;
    /** The app's key. */
    app: string;
    id: string;
}

export interface ComposeImportResult extends SpecImportResult {
    /** The project created, or the one the services went into. */
    projectId?: string;
    /** The apps the services became, as written. */
    apps: ComposeImportApp[];
}

export interface ComposeVariableInput {
    /** Over the .env's; absent keeps that one. */
    value?: string;
    /** Absent decides by the name. */
    secret?: boolean;
}

export interface ComposePortInput {
    published: number;
    target: number;
    protocol: string;
    source?: string;
    as: ComposePortAs;
    domain: string;
}

export interface ComposeServiceInput {
    image?: string;
    /** The app key chosen for it; absent for its name's. */
    app?: string;
    /** Uses the existing env's app its name or key is. */
    useExisting?: boolean;
    ports: ComposePortInput[];
}

/** A file the compose file reads: pasted text, or a file picked. */
export type ComposeFileInput = string | File;

/** Where the services go: a new project, or an env of an existing one. */
export interface ComposeImportProject {
    /** A new project's name; empty takes the file's. Not read for an existing project. */
    name: string;
    /** The env, by its name. */
    env: string;
    /** Creates env in the existing project. */
    newEnv?: boolean;
    envColor?: string;
}

export interface ComposeImportBody {
    /** The existing project the services go into; absent creates one. */
    projectId?: string;
    compose: string;
    dotEnv: string;
    files: Record<string, ComposeFileInput>;
    variables: Record<string, ComposeVariableInput>;
    project: ComposeImportProject;
    profiles: string[];
    services: Record<string, ComposeServiceInput>;
    /** A cluster volume for the services' data; none for the project's own. */
    volumeId?: string;
    selection: SpecImportSelection;
    deploy: boolean;
}
