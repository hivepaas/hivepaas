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
    as: ComposePortAs;
    /** What it becomes unless the review says otherwise. */
    default: ComposePortAs;
    domain: string;
    /** The domain offered under the root domain; empty without one. */
    suggested: string;
}

export type ComposeVolumeKind = "volume" | "shared" | "file" | "host" | "tmpfs" | "dropped";

export interface ComposeVolumeView {
    target: string;
    /** What the file mounts: a volume's name, a path. */
    source: string;
    kind: ComposeVolumeKind;
    readOnly: boolean;
    /** The app whose directory a shared volume is. */
    owner: string;
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
    /** What reads it: env_file, config, secret, bind. */
    as: string;
    by: string[];
    given: boolean;
}

export interface ComposeProject {
    name: string;
    key: string;
    env: string;
    envKey: string;
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

export interface ComposeImportResult extends SpecImportResult {
    /** The project created. */
    projectId?: string;
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
    as: ComposePortAs;
    domain: string;
}

export interface ComposeServiceInput {
    image?: string;
    ports: ComposePortInput[];
}

/** A file the compose file reads: pasted text, or a file picked. */
export type ComposeFileInput = string | File;

export interface ComposeImportBody {
    compose: string;
    dotEnv: string;
    files: Record<string, ComposeFileInput>;
    variables: Record<string, ComposeVariableInput>;
    project: { name: string; env: string };
    profiles: string[];
    services: Record<string, ComposeServiceInput>;
    /** A cluster volume for the services' data; none for the project's own. */
    volumeId?: string;
    selection: SpecImportSelection;
    deploy: boolean;
}
