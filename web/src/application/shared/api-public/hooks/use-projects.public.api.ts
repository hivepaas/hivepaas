import { use, useMemo } from "react";

import { ApplicationPublicApiContext } from "@application/shared/api-public/api-context";
import { type Public_Projects_FindManyPaginated_Req } from "@application/shared/api-public/services";

function createHook() {
    return function useProjectsPublicApi() {
        const { api } = use(ApplicationPublicApiContext);

        const queries = useMemo(
            () => ({
                /**
                 * Find many public projects paginated
                 */
                findManyPaginated: async (
                    data: Public_Projects_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.findManyPaginated(
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

export const useProjectsPublicApi = createHook();
