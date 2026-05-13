import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ImService_CreateOne_Req,
    ImService_CreateOne_Res,
    ImService_DeleteOne_Req,
    ImService_DeleteOne_Res,
    ImService_FindManyPaginated_Req,
    ImService_FindManyPaginated_Res,
    ImService_FindOneById_Req,
    ImService_FindOneById_Res,
    ImService_TestSendMsg_Req,
    ImService_TestSendMsg_Res,
    ImService_UpdateOne_Req,
    ImService_UpdateOne_Res,
    ImService_UpdateStatus_Req,
    ImService_UpdateStatus_Res,
} from "./im-service.api.contracts";
import type { ImServiceApiValidator } from "./im-service.api.validator";

export class ImServiceApi extends BaseApi {
    public constructor(private readonly validator: ImServiceApiValidator) {
        super();
    }

    async findManyPaginated(
        request: ImService_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<ImService_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/im-services", {
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
        request: ImService_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<ImService_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/im-services/${id}`, {
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
        request: ImService_CreateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ImService_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/im-services", payload, {
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
        request: ImService_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<ImService_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/im-services/${id}`, payload, {
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
        request: ImService_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<ImService_UpdateStatus_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/im-services/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: ImService_DeleteOne_Req): Promise<Result<ImService_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/im-services/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async testSendMsg(
        request: ImService_TestSendMsg_Req,
        signal?: AbortSignal,
    ): Promise<Result<ImService_TestSendMsg_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/im-services/test-send-msg", payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.testSendMsg),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
