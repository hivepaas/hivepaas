import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type {
    McpSettings_FindOne_Req,
    McpSettings_FindOne_Res,
    McpSettings_UpdateOne_Req,
    McpSettings_UpdateOne_Res,
} from "./mcp-settings.api.contracts";
import type { McpSettingsApiValidator } from "./mcp-settings.api.validator";

export class McpSettingsApi extends BaseApi {
    public constructor(private readonly validator: McpSettingsApiValidator) {
        super();
    }

    async findOne(
        _request: McpSettings_FindOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<McpSettings_FindOne_Res, Error>> {
        return lastValueFrom(
            from(this.client.v1.get("/system/settings/mcp", { signal })).pipe(
                map(this.validator.findOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }

    async updateOne(
        request: McpSettings_UpdateOne_Req,
        signal?: AbortSignal,
    ): Promise<Result<McpSettings_UpdateOne_Res, Error>> {
        const { payload } = request.data;

        return lastValueFrom(
            from(this.client.v1.put("/system/settings/mcp", payload, { signal })).pipe(
                map(this.validator.updateOne),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
