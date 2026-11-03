import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectImageBuildSettingsApi } from "~/projects/api/hooks";
import type {
    ProjectImageBuildSettings_ClearRepoCache_Req,
    ProjectImageBuildSettings_ClearRepoCache_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type ClearRepoCacheReq = ProjectImageBuildSettings_ClearRepoCache_Req["data"];
type ClearRepoCacheRes = ProjectImageBuildSettings_ClearRepoCache_Res;
type ClearRepoCacheOptions = Omit<UseMutationOptions<ClearRepoCacheRes, Error, ClearRepoCacheReq>, "mutationFn">;

function useClearRepoCache({ onSuccess, ...options }: ClearRepoCacheOptions = {}) {
    const { mutations } = useProjectImageBuildSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.clearRepoCache,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.image-build-settings.repo-cache.$.find-one"], { projectID: request.projectID }],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const ProjectImageBuildSettingsCommands = Object.freeze({
    useClearRepoCache,
});
