import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppConfigFilesApiValidator,
    AppConfigFiles_BuildDownloadUrl_Req,
    AppConfigFiles_CreateOne_Req,
    AppConfigFiles_CreateOne_Res,
    AppConfigFiles_DeleteOne_Req,
    AppConfigFiles_DeleteOne_Res,
    AppConfigFiles_FindManyPaginated_Req,
    AppConfigFiles_FindManyPaginated_Res,
    AppConfigFiles_FindOneById_Req,
    AppConfigFiles_FindOneById_Res,
    AppConfigFiles_GetDownloadToken_Req,
    AppConfigFiles_GetDownloadToken_Res,
    AppConfigFiles_UpdateOne_Req,
    AppConfigFiles_UpdateOne_Res,
} from "~/projects/api/services/project-apps-services";

import { EnvConfig } from "@config";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

export class AppConfigFilesApi extends BaseApi {
    public constructor(private readonly validator: AppConfigFilesApiValidator) {
        super();
    }

    /**
     * Find many app config files paginated
     */
    async findManyPaginated(
        request: AppConfigFiles_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppConfigFiles_FindManyPaginated_Res, Error>> {
        const { projectID, appID, search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/config-files`, {
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
     * Find one app config file by id
     */
    async findOneById(
        request: AppConfigFiles_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppConfigFiles_FindOneById_Res, Error>> {
        const { projectID, appID, configFileID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/config-files/${configFileID}`, {
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
     * Get app config file download token
     */
    async getDownloadToken(
        request: AppConfigFiles_GetDownloadToken_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppConfigFiles_GetDownloadToken_Res, Error>> {
        const { projectID, appID, configFileID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/apps/${appID}/config-files/${configFileID}/download-token`, {
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
     * Build app config file download URL
     */
    buildDownloadUrl(request: AppConfigFiles_BuildDownloadUrl_Req): string {
        const { projectID, appID, configFileID, token, viewInline } = request;
        const baseUrl = EnvConfig.API_URL.endsWith("/") ? EnvConfig.API_URL : `${EnvConfig.API_URL}/`;
        const url = new URL(
            `projects/${encodeURIComponent(projectID)}/apps/${encodeURIComponent(appID)}/config-files/${encodeURIComponent(configFileID)}/download`,
            baseUrl,
        );

        url.searchParams.set("token", token);
        url.searchParams.set("viewInline", String(viewInline));

        return url.toString();
    }

    /**
     * Create an app config file
     */
    async createOne(
        request: AppConfigFiles_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppConfigFiles_CreateOne_Res, Error>> {
        const { projectID, appID, name, content, base64, swarmRef } = request.data;

        const json = {
            name: JsonTransformer.string({
                data: name,
            }),
            content: JsonTransformer.string({
                data: content,
            }),
            base64,
            swarmRef,
        };

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/apps/${appID}/config-files`, json, {
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
     * Delete an app config file
     */
    async deleteOne(request: AppConfigFiles_DeleteOne_Req): Promise<Result<AppConfigFiles_DeleteOne_Res, Error>> {
        const { projectID, appID, configFileID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/apps/${appID}/config-files/${configFileID}`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Update an app config file
     */
    async updateOne(
        request: AppConfigFiles_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppConfigFiles_UpdateOne_Res, Error>> {
        const { projectID, appID, configFileID, updateVer, name, content, base64, swarmRef } = request.data;

        const json = {
            updateVer,
            name: JsonTransformer.string({
                data: name,
            }),
            content: JsonTransformer.string({
                data: content,
            }),
            base64,
            swarmRef,
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/apps/${appID}/config-files/${configFileID}`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
