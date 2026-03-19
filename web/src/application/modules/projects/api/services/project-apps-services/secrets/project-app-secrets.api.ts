import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    AppSecretsApiValidator,
    AppSecrets_CreateOne_Req,
    AppSecrets_CreateOne_Res,
    AppSecrets_DeleteOne_Req,
    AppSecrets_DeleteOne_Res,
    AppSecrets_FindManyPaginated_Req,
    AppSecrets_FindManyPaginated_Res,
    AppSecrets_FindOneById_Req,
    AppSecrets_FindOneById_Res,
    AppSecrets_UpdateOne_Req,
    AppSecrets_UpdateOne_Res,
} from "~/projects/api/services/project-apps-services";
import { EProjectSecretStatus } from "~/projects/module-shared/enums";

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
     * Create an app secret
     */
    async createOne(
        request: AppSecrets_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<AppSecrets_CreateOne_Res, Error>> {
        const { projectID, appID, name, value } = request.data;

        const json = {
            key: JsonTransformer.string({
                data: name,
            }),
            value: JsonTransformer.string({
                data: value,
            }),
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
        const { projectID, appID, secretID, updateVer, name, value } = request.data;

        const json = {
            updateVer,
            name: JsonTransformer.string({
                data: name,
            }),
            value: JsonTransformer.string({
                data: value,
            }),
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
