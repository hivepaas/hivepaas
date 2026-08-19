import { AxiosError, type AxiosResponse, isAxiosError } from "axios";
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    AppContainerFiles_DownloadOne_Req,
    AppContainerFiles_DownloadOne_Res,
    AppContainerFiles_UploadOne_Req,
    AppContainerFiles_UploadOne_Res,
} from "./app-container-files.api.contracts";
import type { AppContainerFilesApiValidator } from "./app-container-files.api.validator";

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

function mapDownloadResponse(response: AxiosResponse<Blob>): AppContainerFiles_DownloadOne_Res {
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

async function parseBlobApiError(error: unknown): Promise<Error> {
    if (!isAxiosError(error) || !(error.response?.data instanceof Blob)) {
        return parseApiError(error);
    }

    try {
        const text = await error.response.data.text();
        const data: unknown = JSON.parse(text);
        const nextError = new AxiosError(error.message, error.code, error.config, error.request, {
            ...error.response,
            data,
        });
        nextError.status = error.status;

        return parseApiError(nextError);
    } catch {
        return parseApiError(error);
    }
}

export class AppContainerFilesApi extends BaseApi {
    constructor(private readonly validator: AppContainerFilesApiValidator) {
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

    async uploadOne(
        req: AppContainerFiles_UploadOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppContainerFiles_UploadOne_Res, Error>> {
        const { projectID, env, appID, nodeId, containerId, path, file, extract, compressionFormat, overwrite } =
            req.data;

        const formData = new FormData();
        formData.append("nodeId", nodeId);
        formData.append("containerId", containerId);
        formData.append("path", path);
        formData.append("extract", String(extract));
        formData.append("compressionFormat", compressionFormat);
        formData.append("overwrite", String(overwrite));
        formData.append("file", file);

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/${env}/apps/${appID}/container/file-upload`, formData, {
                    signal,
                    headers: { "Content-Type": undefined },
                }),
            ).pipe(
                map(this.validator.uploadOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
