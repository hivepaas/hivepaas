import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppImageScanApi } from "../../../api/hooks/project-apps";
import type {
    AppImageScan_CreateIgnore_Req,
    AppImageScan_CreateIgnore_Res,
    AppImageScan_DeleteIgnore_Req,
    AppImageScan_DeleteIgnore_Res,
    AppImageScan_Scan_Req,
    AppImageScan_Scan_Res,
} from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

type Options<Res, Req> = Omit<UseMutationOptions<Res, Error, Req>, "mutationFn">;

/** What an app's image scan is shown from: read again once it changed. */
function useInvalidate() {
    const queryClient = useQueryClient();
    return () => queryClient.invalidateQueries({ queryKey: [QK["projects.apps.image-scan.$.find-one"]] });
}

function useScan({ onSuccess, ...options }: Options<AppImageScan_Scan_Res, AppImageScan_Scan_Req["data"]> = {}) {
    const { mutations } = useAppImageScanApi();
    const invalidate = useInvalidate();

    return useMutation({
        mutationFn: mutations.scan,
        onSuccess: (...args) => {
            void invalidate();
            onSuccess?.(...args);
        },
        ...options,
    });
}

function useCreateIgnore({
    onSuccess,
    ...options
}: Options<AppImageScan_CreateIgnore_Res, AppImageScan_CreateIgnore_Req["data"]> = {}) {
    const { mutations } = useAppImageScanApi();
    const invalidate = useInvalidate();

    return useMutation({
        mutationFn: mutations.createIgnore,
        onSuccess: (...args) => {
            void invalidate();
            onSuccess?.(...args);
        },
        ...options,
    });
}

function useDeleteIgnore({
    onSuccess,
    ...options
}: Options<AppImageScan_DeleteIgnore_Res, AppImageScan_DeleteIgnore_Req["data"]> = {}) {
    const { mutations } = useAppImageScanApi();
    const invalidate = useInvalidate();

    return useMutation({
        mutationFn: mutations.deleteIgnore,
        onSuccess: (...args) => {
            void invalidate();
            onSuccess?.(...args);
        },
        ...options,
    });
}

export const AppImageScanCommands = Object.freeze({
    useScan,
    useCreateIgnore,
    useDeleteIgnore,
});
