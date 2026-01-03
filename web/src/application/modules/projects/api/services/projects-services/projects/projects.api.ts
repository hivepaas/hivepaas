import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import type {
    ProjectsApiValidator,
    Projects_CreateOne_Req,
    Projects_CreateOne_Res,
    Projects_DeleteOne_Req,
    Projects_DeleteOne_Res,
    Projects_FindManyPaginated_Req,
    Projects_FindManyPaginated_Res,
    Projects_FindOneById_Req,
    Projects_FindOneById_Res,
    Projects_UpdateOne_Req,
    Projects_UpdateOne_Res,
} from "~/projects/api/services/projects-services/projects";
import { EProjectStatus } from "~/projects/module-shared/enums";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

export class ProjectsApi extends BaseApi {
    public constructor(private readonly validator: ProjectsApiValidator) {
        super();
    }

    /**
     * Find many projects paginated
     */
    async findManyPaginated(
        request: Projects_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<Projects_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/projects", {
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
     * Find one project by id
     */
    async findOneById(
        request: Projects_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<Projects_FindOneById_Res, Error>> {
        const { projectID } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/projects/${projectID}`, {
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
     * Create a project
     */
    async createOne(request: Projects_CreateOne_Req, signal?: AbortSignal): Promise<Result<Projects_CreateOne_Res, Error>> {
        const { name, note, tags } = request.data;

        const json = {
            name: JsonTransformer.string({
                data: name,
            }),
            note: JsonTransformer.string({
                data: note,
            }),
            tags: JsonTransformer.array({
                data: tags,
            }),
            status: EProjectStatus.Active,
        };
        return lastValueFrom(
            from(this.client.v1.post("/projects", json, { signal })).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Delete a project
     */
    async deleteOne(request: Projects_DeleteOne_Req): Promise<Result<Projects_DeleteOne_Res, Error>> {
        return lastValueFrom(
            from(
                this.client.v1.delete(`/projects/${request.data.projectID}`),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Update a project
     */
    async updateOne(request: Projects_UpdateOne_Req, signal?: AbortSignal): Promise<Result<Projects_UpdateOne_Res, Error>> {
        const { projectID, updateVer, name, note, tags, status } = request.data;

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
                this.client.v1.put(`/projects/${projectID}`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } } as const)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
