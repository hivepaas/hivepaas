import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useProjectStorageSettingsApi } from "../../../api/hooks/project-settings";
import type { ProjectStorageSettings_UpdateOne_Req, ProjectStorageSettings_UpdateOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

type UpdateOneReq = ProjectStorageSettings_UpdateOne_Req["data"];
type UpdateOneRes = ProjectStorageSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectStorageSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.storage-settings.$.find-one"], { projectID: request.projectID }],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const ProjectStorageSettingsCommands = Object.freeze({
    useUpdateOne,
});
