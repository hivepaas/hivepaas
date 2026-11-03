import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { ApplicationApiContext } from "@application/shared/api/api-context";
import type { SettingUsages_FindMany_Req } from "@application/shared/api/services";

function createHook() {
    return function useSettingUsageApi() {
        const { api } = use(ApplicationApiContext);

        const queries = useMemo(
            () => ({
                /**
                 * Deliberately silent about failures.
                 *
                 * The caller is a dialog already explaining a refusal. A second
                 * error on top of it - about the list that explains the first -
                 * would be noise, so the dialog says the list could not be loaded
                 * and leaves the refusal itself standing.
                 */
                findMany: async (data: SettingUsages_FindMany_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settingUsage.findMany({ data }, signal);

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

export const useSettingUsageApi = createHook();
