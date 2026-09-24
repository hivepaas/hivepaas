import type {
    SpecImportOptions,
    SpecImportPlan,
    SpecImportResult,
    SpecImportScope,
    SpecImportSelection,
} from "~/operations/domain";

import type { ApiResponseBase } from "@infrastructure/api";

/**
 * What validate and apply both send. The bundle travels with every call:
 * nothing is kept on the server between them, so a bundle carrying secrets is
 * never stored.
 */
export interface SpecImportBody {
    scope: SpecImportScope;
    /** The file export produced, as it was downloaded. */
    bundle: File;
    /** Opens an encrypted bundle. It is never stored or logged. */
    passphrase?: string;
    selection: SpecImportSelection;
    options: SpecImportOptions;
}

export interface SpecImport_Validate_Req {
    data: SpecImportBody;
}

export type SpecImport_Validate_Res = ApiResponseBase<SpecImportPlan>;

export interface SpecImport_Apply_Req {
    data: SpecImportBody & {
        /** The hash of the plan the operator saw. */
        planHash: string;
        /** Accepts every issue of the plan that is not blocked. */
        acceptIssues: boolean;
    };
}

export type SpecImport_Apply_Res = ApiResponseBase<SpecImportResult>;

/** The body on the wire: the bundle base64 inside the JSON. */
export interface SpecImportWireBody {
    bundle: string;
    passphrase?: string;
    selection: SpecImportSelection;
    options: SpecImportOptions;
    planHash?: string;
    acceptIssues?: boolean;
}
