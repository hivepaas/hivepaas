import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    SSHKey_CreateOne_Req,
    SSHKey_DeleteOne_Req,
    SSHKey_FindManyPaginated_Req,
    SSHKey_FindOneById_Req,
    SSHKey_UpdateMeta_Req,
    SSHKey_UpdateOne_Req,
} from "~/settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSSHKeyApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: SSHKey_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.sshKey.findManyPaginated({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get SSH keys", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: SSHKey_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.sshKey.findOneById({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get SSH key", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                createOne: async (data: SSHKey_CreateOne_Req["data"]) => {
                    const result = await api.settings.sshKey.createOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create SSH key", error });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: SSHKey_UpdateOne_Req["data"]) => {
                    const result = await api.settings.sshKey.updateOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update SSH key", error });
                            throw error;
                        },
                    });
                },
                updateMeta: async (data: SSHKey_UpdateMeta_Req["data"]) => {
                    const result = await api.settings.sshKey.updateMeta({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update SSH key status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: SSHKey_DeleteOne_Req["data"]) => {
                    const result = await api.settings.sshKey.deleteOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete SSH key", error });
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

export const useSSHKeyApi = createHook();
