import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";
import {
    type UsersApiValidator,
    type Users_DeleteOne_Req,
    type Users_DeleteOne_Res,
    type Users_FindManyPaginated_Req,
    type Users_FindManyPaginated_Res,
    type Users_FindOneById_Req,
    type Users_FindOneById_Res,
    type Users_UpdateOne_Req,
    type Users_UpdateOne_Res,
} from "~/user-management/api/services";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

export class UsersApi extends BaseApi {
    public constructor(private readonly validator: UsersApiValidator) {
        super();
    }

    /**
     * Find many users paginated
     */
    async findManyPaginated(
        request: Users_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<Users_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;

        const query = this.queryBuilder.getInstance();

        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/users", {
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
     * Find one user by id
     */
    async findOneById(
        request: Users_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<Users_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/users/${id}`, {
                    signal,
                    params: {
                        getAccesses: true,
                    },
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Delete a user
     */
    async deleteOne(request: Users_DeleteOne_Req): Promise<Result<Users_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/users/${id}`, {})).pipe(
                map(() => Ok({ data: { id } })),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Update a user
     */
    async updateOne(request: Users_UpdateOne_Req, signal?: AbortSignal): Promise<Result<Users_UpdateOne_Res, Error>> {
        const { user } = request.data;

        const json = {
            username: JsonTransformer.string({
                data: user.username,
            }),
            email: JsonTransformer.string({
                data: user.email,
            }),
            fullName: JsonTransformer.string({
                data: user.fullName,
            }),
            position: JsonTransformer.string({
                data: user.position,
            }),
            securityOption: JsonTransformer.string({
                data: user.securityOption,
            }),
            accessExpireAt: JsonTransformer.date({
                data: user.accessExpireAt,
            }),
            projectAccesses: JsonTransformer.array({
                data: user.projectAccesses,
            }),
            moduleAccesses: JsonTransformer.array({
                data: user.moduleAccesses,
            }),
            status: JsonTransformer.string({
                data: user.status,
            }),
            role: JsonTransformer.string({
                data: user.role,
            }),
            notes: JsonTransformer.string({
                data: user.notes,
            }),
        };

        return lastValueFrom(
            from(
                this.client.v1.put(`/users/${user.id}`, json, {
                    signal,
                }),
            ).pipe(
                map(() => Ok({ data: { type: "success" } as const })),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
