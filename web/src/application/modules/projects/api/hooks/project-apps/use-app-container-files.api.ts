import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { useApiErrorNotifications } from "@infrastructure/api";

import type { AppContainerFiles_DownloadOne_Req, AppContainerFiles_UploadOne_Req } from "../../../api/services";
import { ProjectsApiContext } from "../../api-context/projects.api.context";

function createHook() {
    return function useAppContainerFilesApi() {
        const context = use(ProjectsApiContext);
        const { api } = context;
        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                downloadOne: async (request: AppContainerFiles_DownloadOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.containerFiles.$.downloadOne({ data: request }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to download container file", error });
                            throw error;
                        },
                    });
                },
                uploadOne: async (request: AppContainerFiles_UploadOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.containerFiles.$.uploadOne({ data: request }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to upload container file", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { mutations };
    };
}

export const useAppContainerFilesApi = createHook();
