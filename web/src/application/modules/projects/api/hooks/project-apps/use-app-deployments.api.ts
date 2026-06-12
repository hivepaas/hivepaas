import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    AppDeployments_Cancel_Req,
    AppDeployments_FindManyPaginated_Req,
    AppDeployments_FindOneById_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAppDeploymentsApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: AppDeployments_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.deployments.$.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
                findOneById: async (data: AppDeployments_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.deployments.$.findOneById({ data }, signal);

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

        const mutations = useMemo(
            () => ({
                cancel: async (data: AppDeployments_Cancel_Req["data"]) => {
                    const result = await api.projects.apps.deployments.$.cancel({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to cancel app deployment", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return {
            queries,
            mutations,
        };
    };
}

export const useAppDeploymentsApi = createHook();
