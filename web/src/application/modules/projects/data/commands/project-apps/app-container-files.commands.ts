import { type UseMutationOptions, useMutation } from "@tanstack/react-query";

import { useAppContainerFilesApi } from "../../../api/hooks/project-apps";
import type {
    AppContainerFiles_DownloadOne_Req,
    AppContainerFiles_DownloadOne_Res,
    AppContainerFiles_UploadOne_Req,
    AppContainerFiles_UploadOne_Res,
} from "../../../api/services";

type DownloadOneReq = AppContainerFiles_DownloadOne_Req["data"];
type DownloadOneRes = AppContainerFiles_DownloadOne_Res;
type DownloadOneOptions = Omit<UseMutationOptions<DownloadOneRes, Error, DownloadOneReq>, "mutationFn">;

function useDownloadOne(options: DownloadOneOptions = {}) {
    const { mutations } = useAppContainerFilesApi();

    return useMutation({
        mutationFn: (request: DownloadOneReq) => {
            return mutations.downloadOne(request);
        },
        ...options,
    });
}

type UploadOneReq = AppContainerFiles_UploadOne_Req["data"];
type UploadOneRes = AppContainerFiles_UploadOne_Res;
type UploadOneOptions = Omit<UseMutationOptions<UploadOneRes, Error, UploadOneReq>, "mutationFn">;

function useUploadOne(options: UploadOneOptions = {}) {
    const { mutations } = useAppContainerFilesApi();

    return useMutation({
        mutationFn: (request: UploadOneReq) => mutations.uploadOne(request),
        ...options,
    });
}

export const AppContainerFilesCommands = Object.freeze({
    useDownloadOne,
    useUploadOne,
});
