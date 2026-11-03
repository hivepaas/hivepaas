import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useImageBuildSettingsApi } from "~/settings/api/hooks";
import type {
    ImageBuildSettings_ClearRepoCache_Req,
    ImageBuildSettings_ClearRepoCache_Res,
    ImageBuildSettings_UpdateOne_Req,
    ImageBuildSettings_UpdateOne_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type UpdateOneReq = ImageBuildSettings_UpdateOne_Req["data"];
type UpdateOneRes = ImageBuildSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useImageBuildSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.image-build-settings.find-one"]],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type ClearRepoCacheReq = ImageBuildSettings_ClearRepoCache_Req["data"];
type ClearRepoCacheRes = ImageBuildSettings_ClearRepoCache_Res;
type ClearRepoCacheOptions = Omit<UseMutationOptions<ClearRepoCacheRes, Error, ClearRepoCacheReq>, "mutationFn">;

function useClearRepoCache({ onSuccess, ...options }: ClearRepoCacheOptions = {}) {
    const { mutations } = useImageBuildSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.clearRepoCache,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.image-build-settings.repo-cache.find-one"]],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const ImageBuildSettingsCommands = Object.freeze({
    useUpdateOne,
    useClearRepoCache,
});
