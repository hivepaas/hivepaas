import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectSSHKey_CreateOne_Req,
    ProjectSSHKey_DeleteOne_Req,
    ProjectSSHKey_FindManyPaginated_Req,
    ProjectSSHKey_FindOneById_Req,
    ProjectSSHKey_UpdateMeta_Req,
    ProjectSSHKey_UpdateOne_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectSSHKeyApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: ProjectSSHKey_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.sshKey.$.findManyPaginated({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get project SSH keys", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectSSHKey_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.sshKey.$.findOneById({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get project SSH key", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                createOne: async (data: ProjectSSHKey_CreateOne_Req["data"]) => {
                    const result = await api.projects.sshKey.$.createOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create project SSH key", error });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectSSHKey_UpdateOne_Req["data"]) => {
                    const result = await api.projects.sshKey.$.updateOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update project SSH key", error });
                            throw error;
                        },
                    });
                },
                updateMeta: async (data: ProjectSSHKey_UpdateMeta_Req["data"]) => {
                    const result = await api.projects.sshKey.$.updateMeta({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update project SSH key status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectSSHKey_DeleteOne_Req["data"]) => {
                    const result = await api.projects.sshKey.$.deleteOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete project SSH key", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { queries, mutations };
    };
}

export const useProjectSSHKeyApi = createHook();
