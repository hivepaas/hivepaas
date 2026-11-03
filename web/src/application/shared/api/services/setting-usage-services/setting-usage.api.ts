import { Err, Ok, type Result } from "oxide.ts";
import { catchError, from, lastValueFrom, map, of } from "rxjs";

import type { SettingUsages_FindMany_Req, SettingUsages_FindMany_Res } from "@application/shared/api/services";

import { BaseApi, parseApiError } from "@infrastructure/api";

import type { SettingUsageApiValidator } from "./setting-usage.api.validator";

/** What still references a setting. See the contracts for why it takes a URL. */
export class SettingUsageApi extends BaseApi {
    public constructor(private readonly validator: SettingUsageApiValidator) {
        super();
    }

    async findMany(
        request: SettingUsages_FindMany_Req,
        signal?: AbortSignal,
    ): Promise<Result<SettingUsages_FindMany_Res, Error>> {
        const { deleteUrl } = request.data;

        return lastValueFrom(
            from(this.client.v1.get(`${deleteUrl}/usages`, { signal })).pipe(
                map(this.validator.findMany),
                map(res => Ok(res)),
                catchError(error => of(Err(parseApiError(error)))),
            ),
        );
    }
}
