import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { useApiErrorNotifications } from "@infrastructure/api";

import type {
    AppImageScan_CreateIgnore_Req,
    AppImageScan_DeleteIgnore_Req,
    AppImageScan_FindOne_Req,
    AppImageScan_Scan_Req,
} from "../../../api/services";
import { ProjectsApiContext } from "../../api-context/projects.api.context";

function createHook() {
    return function useAppImageScanApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (request: AppImageScan_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.imageScan.$.findOne({ data: request }, signal);
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
                scan: async (request: AppImageScan_Scan_Req["data"]) => {
                    const result = await api.projects.apps.imageScan.$.scan({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to start the scan", error });
                            throw error;
                        },
                    });
                },
                createIgnore: async (request: AppImageScan_CreateIgnore_Req["data"]) => {
                    const result = await api.projects.apps.imageScan.$.createIgnore({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to set the vulnerability aside", error });
                            throw error;
                        },
                    });
                },
                deleteIgnore: async (request: AppImageScan_DeleteIgnore_Req["data"]) => {
                    const result = await api.projects.apps.imageScan.$.deleteIgnore({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to remove the ignore", error });
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

export const useAppImageScanApi = createHook();
