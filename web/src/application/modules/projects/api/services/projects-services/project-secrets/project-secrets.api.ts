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

function getProjectSecretsBasePath(projectID: string, env?: string): string {
    if (env && env !== "all") {
        return `/projects/${projectID}/${encodeURIComponent(env)}/secrets`;
    }

    return `/projects/${projectID}/secrets`;
}

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
        const { projectID, env, search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectSecretsBasePath(projectID, env), {
                    params: query.build(),
                    signal,
                }),
            ).pipe(
                map(response => this.validator.findManyPaginated(response)),
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
        const { projectID, env, secretID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectSecretsBasePath(projectID, env)}/${secretID}`, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.findOneById(response)),
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
        const { projectID, env, name, value, base64 } = request.data;

        const json = {
            key: JsonTransformer.string({
                data: name,
            }),
            value: JsonTransformer.string({
                data: value,
            }),
            base64,
            status: EProjectSecretStatus.Active,
        };

        return lastValueFrom(
            from(
                this.client.v1.post(getProjectSecretsBasePath(projectID, env), json, {
                    signal,
                }),
            ).pipe(
                map(response => this.validator.createOne(response)),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Delete a project secret
     */
    async deleteOne(request: ProjectSecrets_DeleteOne_Req): Promise<Result<ProjectSecrets_DeleteOne_Res, Error>> {
        const { projectID, env, secretID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${getProjectSecretsBasePath(projectID, env)}/${secretID}`)).pipe(
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
        const { projectID, env, secretID, updateVer, name, value, base64 } = request.data;

        const json = {
            updateVer,
            key: JsonTransformer.string({
                data: name,
            }),
            value: JsonTransformer.string({
                data: value,
            }),
            base64,
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectSecretsBasePath(projectID, env)}/${secretID}`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
