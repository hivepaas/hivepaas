import type { ComposeImportBody, ComposeImportResult, ComposeImportReview } from "~/operations/domain";

import type { ApiResponseBase } from "@infrastructure/api";

export interface ComposeImport_Validate_Req {
    data: ComposeImportBody;
}

export type ComposeImport_Validate_Res = ApiResponseBase<ComposeImportReview>;

export interface ComposeImport_Apply_Req {
    data: ComposeImportBody & {
        /** The hash of the plan the operator saw. */
        planHash: string;
        /** Accepts every issue of the plan that is not blocked. */
        acceptIssues: boolean;
    };
}

export type ComposeImport_Apply_Res = ApiResponseBase<ComposeImportResult>;

/** The body on the wire: each file base64 inside the JSON, as Go reads a []byte. */
export interface ComposeImportWireBody {
    compose: string;
    dotEnv: string;
    files: Record<string, string>;
    variables: ComposeImportBody["variables"];
    project: ComposeImportBody["project"];
    profiles: string[];
    services: ComposeImportBody["services"];
    volume?: { id: string };
    selection: ComposeImportBody["selection"];
    deploy: boolean;
    planHash?: string;
    acceptIssues?: boolean;
}
