import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { type ApiHttpResponse, BaseApi, parseApiError, parseBlobApiError } from "@infrastructure/api";

import {
    type AppContainerFiles_DownloadOne_Req,
    type AppContainerFiles_DownloadOne_Res,
    type AppContainerFiles_UploadOne_Req,
    appContainerFilesUploadQuery,
} from "./app-container-files.api.contracts";

function parseFilenameFromContentDisposition(contentDisposition?: string): string | undefined {
    if (!contentDisposition) {
        return undefined;
    }

    const encodedFilename = /filename\*=UTF-8''([^;]+)/i.exec(contentDisposition)?.[1];
    if (encodedFilename) {
        return decodeURIComponent(encodedFilename);
    }

    const filename = /filename="?([^";]+)"?/i.exec(contentDisposition)?.[1];
    return filename ? decodeURIComponent(filename) : undefined;
}

function mapDownloadResponse(response: ApiHttpResponse<Blob>): AppContainerFiles_DownloadOne_Res {
    const headers = response.headers as Record<string, unknown>;
    const contentDisposition = headers["content-disposition"];

    return {
        data: {
            blob: response.data,
            filename: parseFilenameFromContentDisposition(
                typeof contentDisposition === "string" ? contentDisposition : undefined,
            ),
        },
    };
}

export class AppContainerFilesApi extends BaseApi {
    public constructor() {
        super();
    }

    async downloadOne(
        req: AppContainerFiles_DownloadOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppContainerFiles_DownloadOne_Res, Error>> {
        const { projectID, env, appID, nodeId, containerId, path, isDir, compressionFormat } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/container/file-download`, {
                    params: {
                        nodeId,
                        containerId,
                        path,
                        isDir,
                        compressionFormat,
                    },
                    responseType: "blob",
                    signal,
                }),
            ).pipe(
                map(mapDownloadResponse),
                map(res => Ok(res)),
                catchError(error => from(parseBlobApiError(error)).pipe(map(parsed => Err(parsed)))),
            ),
        );
    }

    /**
     * Asks the server whether it would take an upload's stream: it answers what
     * it would refuse the stream with, which a browser cannot read from a refused
     * upgrade. Being a request, it also has an expired session refreshed first.
     */
    async checkUpload(req: AppContainerFiles_UploadOne_Req, signal?: AbortSignal): Promise<Result<void, Error>> {
        const { projectID, env, appID } = req.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/${env}/apps/${appID}/container/file-upload/stream`, {
                    params: appContainerFilesUploadQuery(req.data),
                    signal,
                }),
            ).pipe(
                map(() => Ok(undefined)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
