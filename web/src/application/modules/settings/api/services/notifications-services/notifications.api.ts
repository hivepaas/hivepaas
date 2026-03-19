import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import {
    type NotificationsApiValidator,
} from "./notifications.api.validator";
import {
    type Notifications_CreateOne_Req,
    type Notifications_CreateOne_Res,
    type Notifications_DeleteOne_Req,
    type Notifications_DeleteOne_Res,
    type Notifications_FindManyPaginated_Req,
    type Notifications_FindManyPaginated_Res,
    type Notifications_FindOneById_Req,
    type Notifications_FindOneById_Res,
    type Notifications_UpdateOne_Req,
    type Notifications_UpdateOne_Res,
} from "./notifications.api.contracts";

export class NotificationsApi extends BaseApi {
    public constructor(private readonly validator: NotificationsApiValidator) {
        super();
    }

    /**
     * Find many notifications paginated
     */
    async findManyPaginated(
        request: Notifications_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<Notifications_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/notifications", {
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
     * Find one notification by id
     */
    async findOneById(
        request: Notifications_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<Notifications_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/notifications/${id}`, {
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
     * Create one notification
     */
    async createOne(
        request: Notifications_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<Notifications_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/notifications", payload, {
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
     * Update one notification
     */
    async updateOne(
        request: Notifications_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<Notifications_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/notifications/${id}`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /**
     * Delete one notification
     */
    async deleteOne(request: Notifications_DeleteOne_Req): Promise<Result<Notifications_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/notifications/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
