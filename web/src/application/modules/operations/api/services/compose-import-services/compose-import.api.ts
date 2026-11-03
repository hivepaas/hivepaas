import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of, switchMap } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    ComposeImport_Apply_Req,
    ComposeImport_Apply_Res,
    ComposeImport_Validate_Req,
    ComposeImport_Validate_Res,
} from "./compose-import.api.contracts";
import type { ComposeImportApiMapper } from "./compose-import.api.mapper";
import type { ComposeImportApiValidator } from "./compose-import.api.validator";

/** The endpoints' base: a new project's, or an existing one's. */
function composeBase(projectId?: string): string {
    return projectId ? `/projects/${encodeURIComponent(projectId)}/from-compose` : "/projects/from-compose";
}

export class ComposeImportApi extends BaseApi {
    public constructor(
        private readonly validator: ComposeImportApiValidator,
        private readonly mapper: ComposeImportApiMapper,
    ) {
        super();
    }

    /** Reads a compose file and plans the project it makes. It writes nothing. */
    async validate(
        req: ComposeImport_Validate_Req,
        signal?: AbortSignal,
    ): Promise<Result<ComposeImport_Validate_Res, Error>> {
        return lastValueFrom(
            from(this.mapper.body.toApi(req.data)).pipe(
                switchMap(body =>
                    from(this.client.v1.post(`${composeBase(req.data.projectId)}/validate`, body, { signal })),
                ),
                map(this.validator.validate),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    /** Creates the project the plan validate made of the same body. */
    async apply(req: ComposeImport_Apply_Req, signal?: AbortSignal): Promise<Result<ComposeImport_Apply_Res, Error>> {
        const { planHash, acceptIssues } = req.data;

        return lastValueFrom(
            from(this.mapper.body.toApi(req.data, { planHash, acceptIssues })).pipe(
                switchMap(body =>
                    from(this.client.v1.post(`${composeBase(req.data.projectId)}/apply`, body, { signal })),
                ),
                map(this.validator.apply),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
