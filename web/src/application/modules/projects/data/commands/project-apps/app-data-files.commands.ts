import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAppDataFilesApi } from "~/projects/api/hooks/project-apps";
import type { AppDataFiles_DeleteOne_Req, AppDataFiles_DeleteOne_Res } from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type DeleteOneReq = AppDataFiles_DeleteOne_Req["data"];
type DeleteOneRes = AppDataFiles_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useAppDataFilesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.data-files.$.find-many-paginated"]],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppDataFilesCommands = Object.freeze({
    useDeleteOne,
});
