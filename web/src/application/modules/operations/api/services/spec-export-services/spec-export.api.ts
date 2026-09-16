import type { AxiosResponse } from "axios";
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map } from "rxjs";
import type { SpecExportScope, SpecExportSummary } from "~/operations/domain";

import { BaseApi, parseBlobApiError } from "@infrastructure/api";

import type { SpecExport_Export_Req, SpecExport_Export_Res } from "./spec-export.api.contracts";

/** The counts the server sends alongside the bundle. */
const REPORT_HEADER = "x-hivepaas-spec-report";

export function resolveSpecExportEndpoint(scope: SpecExportScope): string {
    switch (scope.type) {
        case "project":
            return `/projects/${scope.projectID}/spec/export`;
        case "project-env":
            return `/projects/${scope.projectID}/${scope.projectEnvID}/spec/export`;
        case "app":
            return `/projects/${scope.projectID}/${scope.projectEnvID}/apps/${scope.appID}/spec/export`;
        case "global":
        default:
            return "/spec/export";
    }
}

function parseFilenameFromContentDisposition(value: string | undefined): string | undefined {
    if (!value) {
        return undefined;
    }
    const match = /filename\*?=(?:UTF-8'')?"?([^";]+)"?/i.exec(value);

    return match?.[1] ? decodeURIComponent(match[1]) : undefined;
}

/**
 * The summary is read from a response header, which a browser can only see when
 * the server lists it in Access-Control-Expose-Headers. It is listed - but a
 * missing summary is not treated as a failure: the bundle is the point, and
 * report.yaml inside it carries the detail either way.
 */
function parseSummary(raw: unknown): SpecExportSummary | undefined {
    if (typeof raw !== "string" || raw.length === 0) {
        return undefined;
    }
    try {
        return JSON.parse(raw) as SpecExportSummary;
    } catch {
        return undefined;
    }
}

function mapExportResponse(response: AxiosResponse<Blob>): SpecExport_Export_Res {
    const headers = response.headers as Record<string, unknown>;
    const contentDisposition = headers["content-disposition"];

    return {
        data: {
            blob: response.data,
            filename:
                parseFilenameFromContentDisposition(
                    typeof contentDisposition === "string" ? contentDisposition : undefined,
                ) ?? "hivepaas-spec.tar.gz",
            sizeBytes: response.data.size,
            summary: parseSummary(headers[REPORT_HEADER]),
        },
    };
}

export class SpecExportApi extends BaseApi {
    public constructor() {
        super();
    }

    async exportSpec(req: SpecExport_Export_Req, signal?: AbortSignal): Promise<Result<SpecExport_Export_Res, Error>> {
        const { scope, secretsMode, passphrase } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.get(resolveSpecExportEndpoint(scope), {
                    params: { secretsMode, passphrase },
                    responseType: "blob",
                    signal,
                }),
            ).pipe(
                map(mapExportResponse),
                map(res => Ok(res)),
                catchError(error => from(parseBlobApiError(error)).pipe(map(parsed => Err(parsed)))),
            ),
        );
    }
}
