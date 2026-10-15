import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    SystemImageScan_FindApps_Req,
    SystemImageScan_FindOne_Req,
    SystemImageScan_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSystemImageScanApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: SystemImageScan_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.imageScan.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get image scanning settings", error });
                            throw error;
                        },
                    });
                },
                findApps: async (data: SystemImageScan_FindApps_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.imageScan.findApps({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to list the scanned apps", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (data: SystemImageScan_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.imageScan.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update image scanning settings", error });
                            throw error;
                        },
                    });
                },
                execute: async () => {
                    const result = await api.systemSettings.imageScan.execute({ data: {} });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to run the image scan", error });
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

export const useSystemImageScanApi = createHook();
