import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSystemCleanupApi } from "~/system-settings/api/hooks";
import type {
    SystemCleanup_ClearBuildCache_Req,
    SystemCleanup_ClearBuildCache_Res,
    SystemCleanup_ClearRepoCache_Req,
    SystemCleanup_ClearRepoCache_Res,
    SystemCleanup_Execute_Req,
    SystemCleanup_Execute_Res,
    SystemCleanup_UpdateOne_Req,
    SystemCleanup_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = SystemCleanup_UpdateOne_Req["data"];
type UpdateOneRes = SystemCleanup_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;
type ExecuteReq = SystemCleanup_Execute_Req["data"];
type ExecuteRes = SystemCleanup_Execute_Res;
type ExecuteOptions = Omit<UseMutationOptions<ExecuteRes, Error, ExecuteReq>, "mutationFn">;
type ClearRepoCacheReq = SystemCleanup_ClearRepoCache_Req["data"];
type ClearRepoCacheRes = SystemCleanup_ClearRepoCache_Res;
type ClearRepoCacheOptions = Omit<UseMutationOptions<ClearRepoCacheRes, Error, ClearRepoCacheReq>, "mutationFn">;
type ClearBuildCacheReq = SystemCleanup_ClearBuildCache_Req["data"];
type ClearBuildCacheRes = SystemCleanup_ClearBuildCache_Res;
type ClearBuildCacheOptions = Omit<UseMutationOptions<ClearBuildCacheRes, Error, ClearBuildCacheReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSystemCleanupApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.cleanup.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

function useExecute({ onSuccess, ...options }: ExecuteOptions = {}) {
    const { mutations } = useSystemCleanupApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.execute,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.cleanup.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

function useClearRepoCache({ onSuccess, ...options }: ClearRepoCacheOptions = {}) {
    const { mutations } = useSystemCleanupApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.clearRepoCache,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.cleanup.repo-cache.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

function useClearBuildCache(options: ClearBuildCacheOptions = {}) {
    const { mutations } = useSystemCleanupApi();

    return useMutation({
        mutationFn: mutations.clearBuildCache,
        ...options,
    });
}

export const SystemCleanupCommands = Object.freeze({
    useUpdateOne,
    useExecute,
    useClearRepoCache,
    useClearBuildCache,
});
