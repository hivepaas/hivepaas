import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    GithubApp_BeginManifestFlow_Req,
    GithubApp_BeginManifestFlow_Res,
    GithubApp_BeginReprovision_Req,
    GithubApp_BeginReprovision_Res,
    GithubApp_CreateOne_Req,
    GithubApp_CreateOne_Res,
    GithubApp_DeleteOne_Req,
    GithubApp_DeleteOne_Res,
    GithubApp_FindManyPaginated_Req,
    GithubApp_FindManyPaginated_Res,
    GithubApp_FindOneById_Req,
    GithubApp_FindOneById_Res,
    GithubApp_ListInstallations_Req,
    GithubApp_ListInstallations_Res,
    GithubApp_TestConnection_Req,
    GithubApp_TestConnection_Res,
    GithubApp_UpdateOne_Req,
    GithubApp_UpdateOne_Res,
    GithubApp_UpdateStatus_Req,
    GithubApp_UpdateStatus_Res,
} from "./github-app.api.contracts";
import type { GithubAppApiValidator } from "./github-app.api.validator";

export class GithubAppApi extends BaseApi {
    public constructor(private readonly validator: GithubAppApiValidator) {
        super();
    }

    async findManyPaginated(
        request: GithubApp_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<GithubApp_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/github-apps", {
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
        request: GithubApp_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<GithubApp_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/github-apps/${id}`, {
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
        request: GithubApp_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<GithubApp_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/github-apps", payload, {
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
        request: GithubApp_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<GithubApp_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/github-apps/${id}`, payload, {
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
        request: GithubApp_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<GithubApp_UpdateStatus_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/github-apps/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: GithubApp_DeleteOne_Req): Promise<Result<GithubApp_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/github-apps/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async testConnection(
        request: GithubApp_TestConnection_Req,
        signal?: AbortSignal,
    ): Promise<Result<GithubApp_TestConnection_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/github-apps/test-conn", payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.testConnection),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async beginManifestFlow(
        request: GithubApp_BeginManifestFlow_Req,
        signal?: AbortSignal,
    ): Promise<Result<GithubApp_BeginManifestFlow_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/github-apps/manifest-flow/begin", payload, {
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
        request: GithubApp_BeginReprovision_Req,
        signal?: AbortSignal,
    ): Promise<Result<GithubApp_BeginReprovision_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post(`/settings/github-apps/${id}/begin-reprovision`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.beginReprovision),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async listInstallations(
        request: GithubApp_ListInstallations_Req,
        signal?: AbortSignal,
    ): Promise<Result<GithubApp_ListInstallations_Res, Error>> {
        const { payload, pagination, sorting, search } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/github-apps/installations/list", payload, {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(this.validator.listInstallations),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
