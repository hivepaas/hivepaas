import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { type ApiHttpResponse, BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AppImageScan_CreateIgnore_Req,
    AppImageScan_CreateIgnore_Res,
    AppImageScan_DeleteIgnore_Req,
    AppImageScan_DeleteIgnore_Res,
    AppImageScan_DownloadSBOM_Req,
    AppImageScan_DownloadSBOM_Res,
    AppImageScan_FindOne_Req,
    AppImageScan_FindOne_Res,
    AppImageScan_Scan_Req,
    AppImageScan_Scan_Res,
    ProjectImageScan_Find_Req,
    ProjectImageScan_Find_Res,
    VulnIgnoreScope,
} from "./app-image-scan.api.contracts";
import type { AppImageScanApiValidator } from "./app-image-scan.api.validator";

/** Where the ignores of a scope are, seen from an app. */
function ignoresPath(scope: VulnIgnoreScope, projectID: string, env: string, appID: string): string {
    switch (scope) {
        case "app":
            return `/projects/${projectID}/${env}/apps/${appID}/vulnerability-ignores`;
        case "project-env":
            return `/projects/${projectID}/${env}/vulnerability-ignores`;
        case "project":
            return `/projects/${projectID}/vulnerability-ignores`;
        case "global":
            return "/settings/vulnerability-ignores";
    }
}

/** The file name a download is given: what the server names it in Content-Disposition. */
function filenameOf(response: ApiHttpResponse<Blob>): string | undefined {
    const value = (response.headers as Record<string, unknown>)["content-disposition"];
    const filename = typeof value === "string" ? /filename="?([^";]+)"?/i.exec(value)?.[1] : undefined;
    return filename ? decodeURIComponent(filename) : undefined;
}

export class AppImageScanApi extends BaseApi {
    constructor(private readonly validator: AppImageScanApiValidator) {
        super();
    }

    async findOne(
        req: AppImageScan_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppImageScan_FindOne_Res, Error>> {
        const { projectID, env, appID } = req.data;

        return lastValueFrom(
            from(this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/image-scan`, { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findProject(
        req: ProjectImageScan_Find_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectImageScan_Find_Res, Error>> {
        const { projectID } = req.data;

        return lastValueFrom(
            from(this.client.v1.get(`/projects/${projectID}/image-scan`, { signal })).pipe(
                map(this.validator.findProject),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async scan(req: AppImageScan_Scan_Req, signal?: AbortSignal): Promise<Result<AppImageScan_Scan_Res, Error>> {
        const { projectID, env, appID } = req.data;

        return lastValueFrom(
            from(this.client.v1.post(`/projects/${projectID}/${env}/apps/${appID}/image-scan`, {}, { signal })).pipe(
                map(this.validator.scan),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createIgnore(
        req: AppImageScan_CreateIgnore_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppImageScan_CreateIgnore_Res, Error>> {
        const { projectID, env, appID, scope, payload } = req.data;

        return lastValueFrom(
            from(this.client.v1.post(ignoresPath(scope, projectID, env, appID), payload, { signal })).pipe(
                map(this.validator.createIgnore),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteIgnore(
        req: AppImageScan_DeleteIgnore_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppImageScan_DeleteIgnore_Res, Error>> {
        const { projectID, env, appID, scope, ignoreID } = req.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${ignoresPath(scope, projectID, env, appID)}/${ignoreID}`, { signal })).pipe(
                map(this.validator.deleteIgnore),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async downloadSBOM(
        req: AppImageScan_DownloadSBOM_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppImageScan_DownloadSBOM_Res, Error>> {
        const { projectID, env, appID, imageId } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.get<Blob>(`/projects/${projectID}/${env}/apps/${appID}/image-scan/sbom`, {
                    params: { imageId },
                    responseType: "blob",
                    signal,
                }),
            ).pipe(
                map(response => ({ data: { blob: response.data, filename: filenameOf(response) } })),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
