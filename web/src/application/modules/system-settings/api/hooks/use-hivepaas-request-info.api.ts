import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type { HivePaaSRequestInfo_FindOne_Req } from "~/system-settings/api/services";

function createHook() {
    return function useHivePaaSRequestInfoApi() {
        const { api } = use(SystemSettingsApiContext);

        const queries = useMemo(
            () => ({
                /**
                 * Deliberately silent about failures.
                 *
                 * This is a diagnostic shown beside a form field, not something
                 * the operator asked for. If it cannot be read, the field simply
                 * offers no suggestion - interrupting an edit with a toast about
                 * a hint that failed would be worse than the missing hint.
                 */
                findOne: async (data: HivePaaSRequestInfo_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.hivepaasRequestInfo.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api],
        );

        return { queries };
    };
}

export const useHivePaaSRequestInfoApi = createHook();
