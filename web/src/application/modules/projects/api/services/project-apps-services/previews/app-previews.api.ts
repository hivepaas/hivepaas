import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppPreviewsApiValidator,
    AppPreviews_CreateOne_Req,
    AppPreviews_CreateOne_Res,
    AppPreviews_FindManyPaginated_Req,
    AppPreviews_FindManyPaginated_Res,
    AppPreviews_PrepareCreate_Req,
    AppPreviews_PrepareCreate_Res,
} from "~/projects/api/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

export class AppPreviewsApi extends BaseApi {
    public constructor(private readonly validator: AppPreviewsApiValidator) {
        super();
    }

    async findManyPaginated(
        request: AppPreviews_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppPreviews_FindManyPaginated_Res, Error>> {
        const { projectID, appID, search, pagination, sorting, getStats } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/previews`, {
                    params: {
                        ...query.build(),
                        ...(getStats === undefined ? {} : { getStats }),
                    },
                    signal,
                }),
            ).pipe(
                map(this.validator.findManyPaginated),
                map(response => Ok(response)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async prepareCreate(
        request: AppPreviews_PrepareCreate_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppPreviews_PrepareCreate_Res, Error>> {
        const { projectID, appID } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/projects/${projectID}/apps/${appID}/previews/prepare`, {}, { signal })).pipe(
                map(this.validator.prepareCreate),
                map(response => Ok(response)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(request: AppPreviews_CreateOne_Req): Promise<Result<AppPreviews_CreateOne_Res, Error>> {
        const { projectID, appID, repoRef, customSubdomain, noStart } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/apps/${appID}/previews`, {
                    repoRef,
                    customSubdomain,
                    noStart,
                }),
            ).pipe(
                map(this.validator.createOne),
                map(response => Ok(response)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
