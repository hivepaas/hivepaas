import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    ProjectConfigFilesApiValidator,
    ProjectConfigFiles_CreateOne_Req,
    ProjectConfigFiles_CreateOne_Res,
    ProjectConfigFiles_DeleteOne_Req,
    ProjectConfigFiles_DeleteOne_Res,
    ProjectConfigFiles_FindManyPaginated_Req,
    ProjectConfigFiles_FindManyPaginated_Res,
    ProjectConfigFiles_FindOneById_Req,
    ProjectConfigFiles_FindOneById_Res,
    ProjectConfigFiles_UpdateOne_Req,
    ProjectConfigFiles_UpdateOne_Res,
} from "~/projects/api/services/projects-services/project-config-files";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

function getProjectConfigFilesBasePath(projectID: string, env?: string): string {
    if (env && env !== "all") {
        return `/projects/${projectID}/${encodeURIComponent(env)}/config-files`;
    }

    return `/projects/${projectID}/config-files`;
}

export class ProjectConfigFilesApi extends BaseApi {
    public constructor(private readonly validator: ProjectConfigFilesApiValidator) {
        super();
    }

    /**
     * Find many project config files paginated
     */
    async findManyPaginated(
        request: ProjectConfigFiles_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectConfigFiles_FindManyPaginated_Res, Error>> {
        const { projectID, env, search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get(getProjectConfigFilesBasePath(projectID, env), {
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
     * Find one project config file by id
     */
    async findOneById(
        request: ProjectConfigFiles_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectConfigFiles_FindOneById_Res, Error>> {
        const { projectID, env, configFileID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`${getProjectConfigFilesBasePath(projectID, env)}/${configFileID}`, {
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
     * Create a project config file
     */
    async createOne(
        request: ProjectConfigFiles_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectConfigFiles_CreateOne_Res, Error>> {
        const { projectID, env, name, content, base64, inheritable } = request.data;

        const json = {
            name: JsonTransformer.string({
                data: name,
            }),
            content: JsonTransformer.string({
                data: content,
            }),
            base64,
            inheritable,
        };

        return lastValueFrom(
            from(
                this.client.v1.post(getProjectConfigFilesBasePath(projectID, env), json, {
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
     * Delete a project config file
     */
    async deleteOne(
        request: ProjectConfigFiles_DeleteOne_Req,
    ): Promise<Result<ProjectConfigFiles_DeleteOne_Res, Error>> {
        const { projectID, env, configFileID } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`${getProjectConfigFilesBasePath(projectID, env)}/${configFileID}`)).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Update a project config file
     */
    async updateOne(
        request: ProjectConfigFiles_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ProjectConfigFiles_UpdateOne_Res, Error>> {
        const { projectID, env, configFileID, updateVer, name, content, base64, inheritable } = request.data;

        const json = {
            updateVer,
            name: JsonTransformer.string({
                data: name,
            }),
            content: JsonTransformer.string({
                data: content,
            }),
            base64,
            inheritable,
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`${getProjectConfigFilesBasePath(projectID, env)}/${configFileID}`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
