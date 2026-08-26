import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

import type {
    ProjectClusterVolumes_CreateOne_Req,
    ProjectClusterVolumes_CreateOne_Res,
    ProjectClusterVolumes_DeleteOne_Req,
    ProjectClusterVolumes_DeleteOne_Res,
    ProjectClusterVolumes_FindManyPaginated_Req,
    ProjectClusterVolumes_FindManyPaginated_Res,
    ProjectClusterVolumes_FindOneById_Req,
    ProjectClusterVolumes_FindOneById_Res,
    ProjectClusterVolumes_UpdateOne_Req,
    ProjectClusterVolumes_UpdateOne_Res,
    ProjectClusterVolumes_UpdateStatus_Req,
    ProjectClusterVolumes_UpdateStatus_Res,
} from "./project-cluster-volumes.api.contracts";
import type { ProjectClusterVolumesApiValidator } from "./project-cluster-volumes.api.validator";

function getProjectClusterVolumesBasePath(projectID: string, env?: string): string {
    if (env) {
        return `/projects/${projectID}/${encodeURIComponent(env)}/cluster-volumes`;
    }

    return `/projects/${projectID}/cluster-volumes`;
}

export class ProjectClusterVolumesApi extends BaseApi {
    public constructor(private readonly validator: ProjectClusterVolumesApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectClusterVolumes_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectClusterVolumes_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectClusterVolumesBasePath(projectID, env), {
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
        request: ProjectClusterVolumes_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectClusterVolumes_FindOneById_Res, Error>> {
        const { projectID, env, volumeID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectClusterVolumesBasePath(projectID, env)}/${volumeID}`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(
        request: ProjectClusterVolumes_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectClusterVolumes_CreateOne_Res, Error>> {
        const { projectID, env, payload } = request.data;
        const json = {
            ...payload,
            name: JsonTransformer.string({ data: payload.name }),
            driver: JsonTransformer.string({ data: payload.driver }),
            inheritable: false,
        };

        return lastValueFrom(
            from(this.client.v1.post(getProjectClusterVolumesBasePath(projectID, env), json, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectClusterVolumes_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectClusterVolumes_UpdateOne_Res, Error>> {
        const { projectID, env, volumeID, payload } = request.data;
        const json = {
            ...payload,
            inheritable: false,
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectClusterVolumesBasePath(projectID, env)}/${volumeID}`, json, { signal }),
            ).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: ProjectClusterVolumes_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectClusterVolumes_UpdateStatus_Res, Error>> {
        const { projectID, env, volumeID, payload } = request.data;
        const json = {
            ...payload,
            inheritable: false,
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectClusterVolumesBasePath(projectID, env)}/${volumeID}/status`, json, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(
        request: ProjectClusterVolumes_DeleteOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectClusterVolumes_DeleteOne_Res, Error>> {
        const { projectID, env, volumeID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.delete(`${getProjectClusterVolumesBasePath(projectID, env)}/${volumeID}`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
