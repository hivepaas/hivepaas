import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    ProjectAppsApiValidator,
    ProjectApps_Copy_Req,
    ProjectApps_Copy_Res,
    ProjectApps_CreateOne_Req,
    ProjectApps_CreateOne_Res,
    ProjectApps_DeleteOne_Req,
    ProjectApps_DeleteOne_Res,
    ProjectApps_Deploy_Req,
    ProjectApps_Deploy_Res,
    ProjectApps_FindManyPaginated_Req,
    ProjectApps_FindManyPaginated_Res,
    ProjectApps_FindOneById_Req,
    ProjectApps_FindOneById_Res,
    ProjectApps_PrepareCopy_Req,
    ProjectApps_PrepareCopy_Res,
    ProjectApps_Restart_Req,
    ProjectApps_Restart_Res,
    ProjectApps_UpdateOne_Req,
    ProjectApps_UpdateOne_Res,
    ProjectApps_UpdateStatus_Req,
    ProjectApps_UpdateStatus_Res,
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
        const { projectID, search, pagination, sorting, env, getStats } = request.data;

        const query = this.queryBuilder.getInstance();

        query
            .pagination(pagination)
            .sorting(sorting)
            .search(search)
            .filterBy({ env: [env] });

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps`, {
                    params: {
                        ...query.build(),
                        ...(getStats === undefined ? {} : { getStats }),
                    },
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

    async prepareCopy(
        request: ProjectApps_PrepareCopy_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectApps_PrepareCopy_Res, Error>> {
        const { projectID, appID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/copy/prepare`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.prepareCopy),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async copy(request: ProjectApps_Copy_Req, signal?: AbortSignal): Promise<Result<ProjectApps_Copy_Res, Error>> {
        const { projectID, appID, ...json } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/apps/${appID}/copy`, json, {
                    signal,
                }),
            ).pipe(
                map(this.validator.copy),
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
        const { projectID, name, env, note, tags } = request.data;

        const json = {
            name: JsonTransformer.string({
                data: name,
            }),
            env: JsonTransformer.string({
                data: env,
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

    /**
     * Update a project app status
     */
    async updateStatus(
        request: ProjectApps_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectApps_UpdateStatus_Res, Error>> {
        const { projectID, appID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Deploy a project app using existing deployment settings
     */
    async deploy(request: ProjectApps_Deploy_Req): Promise<Result<ProjectApps_Deploy_Res, Error>> {
        const { projectID, appID } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/projects/${projectID}/apps/${appID}/deploy`, {})).pipe(
                map(this.validator.deploy),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Restart a project app
     */
    async restart(request: ProjectApps_Restart_Req): Promise<Result<ProjectApps_Restart_Res, Error>> {
        const { projectID, appID } = request.data;

        return lastValueFrom(
            from(this.client.v1.post(`/projects/${projectID}/apps/${appID}/restart`, {})).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
