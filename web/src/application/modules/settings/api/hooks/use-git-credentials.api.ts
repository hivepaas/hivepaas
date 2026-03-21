import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type { GitCredentials_FindManyPaginated_Req } from "~/settings/api/services/git-credentials-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useGitCredentialsApi() {
        const { api } = use(SettingsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: GitCredentials_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.gitCredentials.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get git credentials",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return {
            queries,
        };
    };
}

export const useGitCredentialsApi = createHook();
