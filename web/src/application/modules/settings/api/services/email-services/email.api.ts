import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    Email_CreateOne_Req,
    Email_CreateOne_Res,
    Email_DeleteOne_Req,
    Email_DeleteOne_Res,
    Email_FindManyPaginated_Req,
    Email_FindManyPaginated_Res,
    Email_FindOneById_Req,
    Email_FindOneById_Res,
    Email_TestSendMail_Req,
    Email_TestSendMail_Res,
    Email_UpdateOne_Req,
    Email_UpdateOne_Res,
    Email_UpdateStatus_Req,
    Email_UpdateStatus_Res,
} from "./email.api.contracts";
import type { EmailApiValidator } from "./email.api.validator";

export class EmailApi extends BaseApi {
    public constructor(private readonly validator: EmailApiValidator) {
        super();
    }

    async findManyPaginated(
        request: Email_FindManyPaginated_Req,
        signal?: AbortSignal,
    ): Promise<Result<Email_FindManyPaginated_Res, Error>> {
        const { search, pagination, sorting } = request.data;
        const query = this.queryBuilder.getInstance();
        query.pagination(pagination).sorting(sorting).search(search);

        return lastValueFrom(
            from(
                this.client.v1.get("/settings/emails", {
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
        request: Email_FindOneById_Req,
        signal?: AbortSignal,
    ): Promise<Result<Email_FindOneById_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.get(`/settings/emails/${id}`, {
                    signal,
                }),
            ).pipe(
                map(this.validator.findOneById),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async createOne(request: Email_CreateOne_Req, signal?: AbortSignal): Promise<Result<Email_CreateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/emails", payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(request: Email_UpdateOne_Req, signal?: AbortSignal): Promise<Result<Email_UpdateOne_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/emails/${id}`, payload, {
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
        request: Email_UpdateStatus_Req,
        signal?: AbortSignal,
    ): Promise<Result<Email_UpdateStatus_Res, Error>> {
        const { id, payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.put(`/settings/emails/${id}/status`, payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.updateStatus),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async deleteOne(request: Email_DeleteOne_Req): Promise<Result<Email_DeleteOne_Res, Error>> {
        const { id } = request.data;

        return lastValueFrom(
            from(this.client.v1.delete(`/settings/emails/${id}`)).pipe(
                map(this.validator.deleteOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async testSendMail(
        request: Email_TestSendMail_Req,
        signal?: AbortSignal,
    ): Promise<Result<Email_TestSendMail_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(
                this.client.v1.post("/settings/emails/test-send-mail", payload, {
                    signal,
                }),
            ).pipe(
                map(this.validator.testSendMail),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
