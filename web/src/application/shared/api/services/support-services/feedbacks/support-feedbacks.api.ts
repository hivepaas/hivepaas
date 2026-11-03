import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, JsonTransformer, parseApiError } from "@infrastructure/api";

import type { SupportFeedbacks_CreateOne_Req, SupportFeedbacks_CreateOne_Res } from "./support-feedbacks.api.contracts";
import type { SupportFeedbacksApiValidator } from "./support-feedbacks.api.validator";

export class SupportFeedbacksApi extends BaseApi {
    public constructor(private readonly validator: SupportFeedbacksApiValidator) {
        super();
    }

    async createOne(request: SupportFeedbacks_CreateOne_Req): Promise<Result<SupportFeedbacks_CreateOne_Res, Error>> {
        const { data } = request;

        return lastValueFrom(
            from(
                this.client.v1.post("/support/feedbacks", {
                    category: data.category,
                    name: JsonTransformer.string({ data: data.name }),
                    email: JsonTransformer.string({ data: data.email }),
                    company: JsonTransformer.string({ data: data.company }),
                    subject: data.subject,
                    description: data.description,
                }),
            ).pipe(
                map(this.validator.createOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
