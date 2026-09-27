import { use, useMemo } from "react";

import { ApplicationPublicApiContext } from "@application/shared/api-public/api-context";
import { type Public_Apps_FindMany_Req } from "@application/shared/api-public/services";

function createHook() {
    return function useAppsPublicApi() {
        const { api } = use(ApplicationPublicApiContext);

        const queries = useMemo(
            () => ({
                /**
                 * Find many public apps
                 */
                findMany: async (data: Public_Apps_FindMany_Req["data"], signal?: AbortSignal) => {
                    const result = await api.apps.findMany(
                        {
                            data,
                        },
                        signal,
                    );

                    return result.unwrap();
                },
            }),
            [api],
        );

        return {
            queries,
        };
    };
}

export const useAppsPublicApi = createHook();
