import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ProjectGithubApp_BeginManifestFlow_Req,
    ProjectGithubApp_BeginManifestFlow_Res,
    ProjectGithubApp_BeginReprovision_Req,
    ProjectGithubApp_BeginReprovision_Res,
    ProjectGithubApp_CreateOne_Req,
    ProjectGithubApp_CreateOne_Res,
    ProjectGithubApp_DeleteOne_Req,
    ProjectGithubApp_DeleteOne_Res,
    ProjectGithubApp_FindManyPaginated_Req,
    ProjectGithubApp_FindManyPaginated_Res,
    ProjectGithubApp_FindOneById_Req,
    ProjectGithubApp_FindOneById_Res,
    ProjectGithubApp_UpdateOne_Req,
    ProjectGithubApp_UpdateOne_Res,
    ProjectGithubApp_UpdateStatus_Req,
    ProjectGithubApp_UpdateStatus_Res,
} from "./project-github-app.api.contracts";
import type { ProjectGithubAppApiValidator } from "./project-github-app.api.validator";

export class ProjectGithubAppApi extends BaseApi {
    public constructor(private readonly validator: ProjectGithubAppApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ProjectGithubApp_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGithubApp_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/github-apps`, {
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
        request: ProjectGithubApp_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGithubApp_FindOneById_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/github-apps/${id}`, {
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
        request: ProjectGithubApp_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGithubApp_CreateOne_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/github-apps`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: ProjectGithubApp_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGithubApp_UpdateOne_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/github-apps/${id}`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateStatus(
        request: ProjectGithubApp_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGithubApp_UpdateStatus_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/github-apps/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ProjectGithubApp_DeleteOne_Req): Promise<Result<ProjectGithubApp_DeleteOne_Res, Error>> {
        const { projectID, id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/github-apps/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async beginManifestFlow(
        request: ProjectGithubApp_BeginManifestFlow_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGithubApp_BeginManifestFlow_Res, Error>> {
        const { projectID, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/github-apps/manifest-flow/begin`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.beginManifestFlow),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async beginReprovision(
        request: ProjectGithubApp_BeginReprovision_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectGithubApp_BeginReprovision_Res, Error>> {
        const { projectID, id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/github-apps/${id}/begin-reprovision`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.beginReprovision),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
