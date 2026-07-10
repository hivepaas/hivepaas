import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppDataFilesApiValidator,
    AppDataFiles_DeleteOne_Req,
    AppDataFiles_DeleteOne_Res,
    AppDataFiles_FindManyPaginated_Req,
    AppDataFiles_FindManyPaginated_Res,
    AppDataFiles_GetDownloadUrl_Req,
    AppDataFiles_GetDownloadUrl_Res,
} from "~/projects/api/services/project-apps-services/data-files";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class AppDataFilesApi extends BaseApi {
    public constructor(private readonly validator: AppDataFilesApiValidator) {
        super();
    }

    async findManyPaginated(
        request: AppDataFiles_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppDataFiles_FindManyPaginated_Res, Error>> {
        const { projectID, appID, pagination, sorting, search } = request.data;
        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/data-files`, {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyPaginated),
                map(response => Ok(response)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async getDownloadUrl(
        request: AppDataFiles_GetDownloadUrl_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppDataFiles_GetDownloadUrl_Res, Error>> {
        const { projectID, appID, dataFileID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/data-files/${dataFileID}/download-url`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.getDownloadUrl),
                map(response => Ok(response)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: AppDataFiles_DeleteOne_Req): Promise<Result<AppDataFiles_DeleteOne_Res, Error>> {
        const { projectID, appID, dataFileID, deletePermanently } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.delete(`/projects/${projectID}/apps/${appID}/data-files/${dataFileID}`, {
                    params: { deletePermanently },
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
