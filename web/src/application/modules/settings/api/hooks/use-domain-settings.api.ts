import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    DomainSettings_DeleteOne_Req,
    DomainSettings_FindOne_Req,
    DomainSettings_UpdateOne_Req,
    DomainSettings_UpdateStatus_Req,
} from "~/settings/api/services/domain-settings-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useDomainSettingsApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: DomainSettings_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.domainSettings.findOne(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get domain settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (data: DomainSettings_UpdateOne_Req["data"]) => {
                    const result = await api.settings.domainSettings.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update domain settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: DomainSettings_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.domainSettings.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update domain settings status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: DomainSettings_DeleteOne_Req["data"]) => {
                    const result = await api.settings.domainSettings.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete domain settings",
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
            mutations,
        };
    };
}

export const useDomainSettingsApi = createHook();
