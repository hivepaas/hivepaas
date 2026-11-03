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
    /** Where the dump goes: empty for the archive's repository, if it is one; NO_DUMP for none; a repository's id. */
    dumpRepo: string;
    dumpSchedule: string;
}

export const EMPTY_MANAGED_FORM: ManagedFormState = {
    archive: "",
    keyAuth: "",
    archiveTimeout: "",
    retainFull: "",
    baseBackupSchedule: "",
    restoreTestSchedule: "",
    restoreTestQuery: "",
    dumpRepo: "",
    dumpSchedule: "",
};

/** The key select's value for the storage's own key. */
export const NO_KEY = "__storage-key__";

/** The dump select's value for no dumps. */
export const NO_DUMP = "__no-dump__";

/** What the create request says of a managed app: a field left empty is the template's default, then the engine's. */
export function managedCreateRequest(state: ManagedFormState): ManagedCreateReq {
    const [kind, id = ""] = state.archive.split(":");
    const keyAuth = state.keyAuth !== "" && state.keyAuth !== NO_KEY ? { id: state.keyAuth } : undefined;
    const retainFull = Number.parseInt(state.retainFull, 10);
    let dumpRepo: { id: string } | null = null;
    if (state.dumpRepo !== "" && state.dumpRepo !== NO_DUMP) {
        dumpRepo = { id: state.dumpRepo };
    } else if (state.dumpRepo === "" && kind === "repo") {
        dumpRepo = { id };
    }
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
        dump: { repo: dumpRepo, schedule: state.dumpSchedule.trim() || undefined },
    };
}
