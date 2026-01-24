import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    ProjectSecretsApiValidator,
    ProjectSecrets_CreateOne_Req,
    ProjectSecrets_CreateOne_Res,
    ProjectSecrets_DeleteOne_Req,
    ProjectSecrets_DeleteOne_Res,
    ProjectSecrets_FindManyPaginated_Req,
    ProjectSecrets_FindManyPaginated_Res,
    ProjectSecrets_FindOneById_Req,
    ProjectSecrets_FindOneById_Res,
    ProjectSecrets_UpdateOne_Req,
    ProjectSecrets_UpdateOne_Res,
} from "~/projects/api/services/projects-services/project-secrets";
import { EProjectSecretStatus } from "~/projects/module-shared/enums";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

export class ProjectSecretsApi extends BaseApi {
    public constructor(private readonly validator: ProjectSecretsApiValidator) {
        super();
    }

    /**
     * Find many project secrets paginated
     */
    async findManyPaginated(
        request: ProjectSecrets_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSecrets_FindManyPaginated_Res, Error>> {
        const { projectID, search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/secrets`, {
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
     * Find one project secret by id
     */
    async findOneById(
        request: ProjectSecrets_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSecrets_FindOneById_Res, Error>> {
        const { projectID, secretID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}/secrets/${secretID}`, {
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
     * Create a project secret
     */
    async createOne(
        request: ProjectSecrets_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSecrets_CreateOne_Res, Error>> {
        const { projectID, name, key } = request.data;

        const json = {
            name: JsonTransformer.string({
                data: name,
            }),
            key: JsonTransformer.string({
                data: key,
            }),
            status: EProjectSecretStatus.Active,
        };

        return lastValueFrom(
            from(
                this.client.v1.post(`/projects/${projectID}/secrets`, json, {
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
     * Delete a project secret
     */
    async deleteOne(request: ProjectSecrets_DeleteOne_Req): Promise<Result<ProjectSecrets_DeleteOne_Res, Error>> {
        const { projectID, secretID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/projects/${projectID}/secrets/${secretID}`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Update a project secret
     */
    async updateOne(
        request: ProjectSecrets_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectSecrets_UpdateOne_Res, Error>> {
        const { projectID, secretID, updateVer, name } = request.data;

        const json = {
            updateVer,
            name: JsonTransformer.string({
                data: name,
            }),
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`/projects/${projectID}/secrets/${secretID}`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
