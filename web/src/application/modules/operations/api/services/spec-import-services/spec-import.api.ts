import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of, switchMap } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    SpecImport_Apply_Req,
    SpecImport_Apply_Res,
    SpecImport_Validate_Req,
    SpecImport_Validate_Res,
} from "./spec-import.api.contracts";
import { type SpecImportApiMapper, resolveSpecImportEndpoint } from "./spec-import.api.mapper";
import type { SpecImportApiValidator } from "./spec-import.api.validator";

export class SpecImportApi extends BaseApi {
    public constructor(
        private readonly validator: SpecImportApiValidator,
        private readonly mapper: SpecImportApiMapper,
    ) {
        super();
    }

    /** Plans an import. It writes nothing. */
    async validate(
        req: SpecImport_Validate_Req,
        signal?: AbortSignal,
    ): Promise<Result<SpecImport_Validate_Res, Error>> {
        return lastValueFrom(
            from(this.mapper.body.toApi(req.data)).pipe(
                switchMap(body =>
                    from(this.client.v1.post(resolveSpecImportEndpoint(req.data.scope, "validate"), body, { signal })),
                ),
                map(this.validator.validate),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /** Applies the plan validate made of the same body. */
    async apply(req: SpecImport_Apply_Req, signal?: AbortSignal): Promise<Result<SpecImport_Apply_Res, Error>> {
        const { planHash, acceptIssues } = req.data;

        return lastValueFrom(
            from(this.mapper.body.toApi(req.data, { planHash, acceptIssues })).pipe(
                switchMap(body =>
                    from(this.client.v1.post(resolveSpecImportEndpoint(req.data.scope, "apply"), body, { signal })),
                ),
                map(this.validator.apply),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
