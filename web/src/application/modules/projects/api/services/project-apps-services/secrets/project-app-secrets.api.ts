import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppSecretsApiValidator,
    AppSecrets_BuildDownloadUrl_Req,
    AppSecrets_CreateOne_Req,
    AppSecrets_CreateOne_Res,
    AppSecrets_DeleteOne_Req,
    AppSecrets_DeleteOne_Res,
    AppSecrets_FindManyPaginated_Req,
    AppSecrets_FindManyPaginated_Res,
    AppSecrets_FindOneById_Req,
    AppSecrets_FindOneById_Res,
    AppSecrets_GetDownloadToken_Req,
    AppSecrets_GetDownloadToken_Res,
    AppSecrets_UpdateOne_Req,
    AppSecrets_UpdateOne_Res,
} from "~/projects/api/services/project-apps-services";
import { EProjectSecretStatus } from "~/projects/module-shared/enums";

import { EnvConfig } from "@config";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

export class AppSecretsApi extends BaseApi {
    public constructor(private readonly validator: AppSecretsApiValidator) {
        super();
    }

    /**
     * Find many app secrets paginated
     */
    async findManyPaginated(
        request: AppSecrets_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSecrets_FindManyPaginated_Res, Error>> {
        const { projectID, appID, search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/secrets`, {
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
     * Find one app secret by id
     */
    async findOneById(
        request: AppSecrets_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSecrets_FindOneById_Res, Error>> {
        const { projectID, appID, secretID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/secrets/${secretID}`, {
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
     * Get app secret download token
     */
    async getDownloadToken(
        request: AppSecrets_GetDownloadToken_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSecrets_GetDownloadToken_Res, Error>> {
        const { projectID, appID, secretID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/secrets/${secretID}/download-token`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.getDownloadToken),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Build app secret download URL
     */
    buildDownloadUrl(request: AppSecrets_BuildDownloadUrl_Req): string {
        const { projectID, appID, secretID, token, viewInline } = request;
        const baseUrl = EnvConfig.API_URL.endsWith("/") ? EnvConfig.API_URL : `${EnvConfig.API_URL}/`;
        const url = new URL(
            `projects/${encodeURIComponent(projectID)}/apps/${encodeURIComponent(appID)}/secrets/${encodeURIComponent(secretID)}/download`,
            baseUrl,
        );

        url.searchParams.set("token", token);
        url.searchParams.set("viewInline", String(viewInline));

        return url.toString();
    }

    /**
     * Create an app secret
     */
    async createOne(
        request: AppSecrets_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSecrets_CreateOne_Res, Error>> {
        const { projectID, appID, name, value, base64, swarmRef } = request.data;

        const json = {
            key: JsonTransformer.string({
                data: name,
            }),
            value: JsonTransformer.string({
                data: value,
            }),
            base64,
            swarmRef,
            status: EProjectSecretStatus.Active,
        };

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/apps/${appID}/secrets`, json, {
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
     * Delete an app secret
     */
    async deleteOne(request: AppSecrets_DeleteOne_Req): Promise<Result<AppSecrets_DeleteOne_Res, Error>> {
        const { projectID, appID, secretID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/apps/${appID}/secrets/${secretID}`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Update an app secret
     */
    async updateOne(
        request: AppSecrets_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSecrets_UpdateOne_Res, Error>> {
        const { projectID, appID, secretID, updateVer, name, value, base64, swarmRef } = request.data;

        const json = {
            updateVer,
            key: JsonTransformer.string({
                data: name,
            }),
            value: JsonTransformer.string({
                data: value,
            }),
            base64,
            swarmRef,
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/secrets/${secretID}`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
