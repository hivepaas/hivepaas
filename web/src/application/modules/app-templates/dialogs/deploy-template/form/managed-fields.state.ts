import type { ManagedCreateReq } from "../../../api";

/** A place a managed database can archive to: a backup repository, on S3 or a volume, or a cloud storage. */
export interface ManagedArchiveOption {
    id: string;
    name: string;
    /** "repo" for a backup repository, "storage" for a cloud storage. */
    kind: "repo" | "storage";
    /** A backup repository whose storage is a volume rather than object storage. */
    onVolume: boolean;
}

export interface ManagedFormState {
    /** The archive, as `<kind>:<id>`. */
    archive: string;
    keyAuth: string;
    archiveTimeout: string;
    retainFull: string;
    baseBackupSchedule: string;
    restoreTestSchedule: string;
    restoreTestQuery: string;
}

export const EMPTY_MANAGED_FORM: ManagedFormState = {
    archive: "",
    keyAuth: "",
    archiveTimeout: "",
    retainFull: "",
    baseBackupSchedule: "",
    restoreTestSchedule: "",
    restoreTestQuery: "",
};

/** The key select's value for the storage's own key. */
export const NO_KEY = "__storage-key__";

/** What the create request says of a managed app: a field left empty is the template's default, then the engine's. */
export function managedCreateRequest(state: ManagedFormState): ManagedCreateReq {
    const [kind, id = ""] = state.archive.split(":");
    const keyAuth = state.keyAuth !== "" && state.keyAuth !== NO_KEY ? { id: state.keyAuth } : undefined;
    const retainFull = Number.parseInt(state.retainFull, 10);
    return {
        storage: {
            backupRepo: kind === "repo" ? { id } : undefined,
            cloudStorage: kind === "storage" ? { id } : undefined,
            keyAuth,
        },
        archiveTimeout: state.archiveTimeout.trim() || undefined,
        retainFull: Number.isNaN(retainFull) ? undefined : retainFull,
        baseBackupSchedule: state.baseBackupSchedule.trim() || undefined,
        restoreTestSchedule: state.restoreTestSchedule.trim() || undefined,
        restoreTestQuery: state.restoreTestQuery.trim() || undefined,
    };
}
