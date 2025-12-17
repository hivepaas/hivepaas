import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { ApplicationApiContext } from "@application/shared/api/api-context";

import type { SshKeys_FindManyPaginated_Req } from "../../services";

function createHook() {
    return function useSshKeysApi() {
        const { api } = use(ApplicationApiContext);


        const queries = useMemo(
            () => ({
                /**
                 * Find many SSH keys paginated
                 */
                findManyPaginated: async (data: SshKeys_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.providers.sshKeys.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

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

        const mutations = useMemo(() => ({}), []);

        return {
            queries,
            mutations,
        };
    };
}

export const useSshKeysApi = createHook();
