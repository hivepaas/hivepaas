import type { AxiosResponse } from "axios";
import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SystemBackupFile_DeleteOne_Req,
    SystemBackupFile_DeleteOne_Res,
    SystemBackupFile_DownloadOne_Req,
    SystemBackupFile_DownloadOne_Res,
    SystemBackupFile_FindManyPaginated_Req,
    SystemBackupFile_FindManyPaginated_Res,
    SystemBackupFile_FindOneById_Req,
    SystemBackupFile_FindOneById_Res,
} from "./system-backup-file.api.contracts";
import type { SystemBackupFileApiValidator } from "./system-backup-file.api.validator";

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

function mapDownloadResponse(response: AxiosResponse<Blob>): SystemBackupFile_DownloadOne_Res {
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

export class SystemBackupFileApi extends BaseApi {
    public constructor(private readonly validator: SystemBackupFileApiValidator) {
        super();
    }

    async findManyPaginated(
        request: SystemBackupFile_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemBackupFile_FindManyPaginated_Res, Error>> {
        const { pagination, sorting, search } = request.data;
        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/system/settings/backup/files", {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyPaginated),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async findOneById(
        request: SystemBackupFile_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemBackupFile_FindOneById_Res, Error>> {
        const { fileID } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`/system/settings/backup/files/${fileID}`, { signal })).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async downloadOne(
        request: SystemBackupFile_DownloadOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<SystemBackupFile_DownloadOne_Res, Error>> {
        const { fileID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get<Blob>(`/system/settings/backup/files/${fileID}/download`, {
                    responseType: "blob",
                    signal,
                }),
            ).pipe(
                map(mapDownloadResponse),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: SystemBackupFile_DeleteOne_Req): Promise<Result<SystemBackupFile_DeleteOne_Res, Error>> {
        const { fileID, deletePermanently } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.delete(`/system/settings/backup/files/${fileID}`, {
                    params: { deletePermanently },
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
