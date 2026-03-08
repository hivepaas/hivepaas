import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    ProjectAppsApiValidator,
    ProjectApps_CreateOne_Req,
    ProjectApps_CreateOne_Res,
    ProjectApps_DeleteOne_Req,
    ProjectApps_DeleteOne_Res,
    ProjectApps_FindManyPaginated_Req,
    ProjectApps_FindManyPaginated_Res,
    ProjectApps_FindOneById_Req,
    ProjectApps_FindOneById_Res,
    ProjectApps_UpdateOne_Req,
    ProjectApps_UpdateOne_Res,
} from "~/projects/api/services";
import { EProjectAppStatus } from "~/projects/module-shared/enums";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

export class ProjectAppsApi extends BaseApi {
    public constructor(private readonly validator: ProjectAppsApiValidator) {
        super();
    }

    /**
     * Find many project apps paginated
     */
    async findManyPaginated(
        request: ProjectApps_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectApps_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps`, {
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

    /**
     * Find one project app by id
     */
    async findOneById(
        request: ProjectApps_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectApps_FindOneById_Res, Error>> {
        const { projectID, appID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Create a project app
     */
    async createOne(
        request: ProjectApps_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectApps_CreateOne_Res, Error>> {
        const { projectID, name, note, tags } = request.data;

        const json = {
            name: JsonTransformer.string({
                data: name,
            }),
            note: JsonTransformer.string({
                data: note,
            }),
            tags: JsonTransformer.array({
                data: tags,
            }),
            status: EProjectAppStatus.Active,
        };

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/apps`, json, {
                    signal,
                }),
            ).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Delete a project app
     */
    async deleteOne(request: ProjectApps_DeleteOne_Req): Promise<Result<ProjectApps_DeleteOne_Res, Error>> {
        const { projectID, appID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/apps/${appID}`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Update a project app
     */
    async updateOne(
        request: ProjectApps_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectApps_UpdateOne_Res, Error>> {
        const { projectID, appID, updateVer, name, note, tags, status } = request.data;

        const json: Record<string, unknown> = {
            updateVer,
            name: JsonTransformer.string({
                data: name,
            }),
            note: JsonTransformer.string({
                data: note,
            }),
            tags: JsonTransformer.array({
                data: tags,
            }),
            status: JsonTransformer.string({
                data: status,
            }),
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
