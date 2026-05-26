import { use, useMemo } from "react";

import { ApplicationPublicApiContext } from "@application/shared/api-public/api-context";
import { type Public_Users_FindManyBase_Req } from "@application/shared/api-public/services";

function createHook() {
    return function useUsersPublicApi() {
        const { api } = use(ApplicationPublicApiContext);

        const queries = useMemo(
            () => ({
                /**
                 * Find many public users base
                 */
                findManyBase: async (data: Public_Users_FindManyBase_Req["data"], signal?: AbortSignal) => {
                    const result = await api.users.findManyBase(
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

export const useUsersPublicApi = createHook();
