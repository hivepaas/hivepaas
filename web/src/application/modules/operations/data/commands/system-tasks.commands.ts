import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSystemTasksApi } from "~/operations/api";
import type { SystemTasks_Cancel_Req, SystemTasks_Cancel_Res } from "~/operations/api/services";
import { QK } from "~/operations/data/constants";

type CancelReq = SystemTasks_Cancel_Req["data"];
type CancelRes = SystemTasks_Cancel_Res;
type CancelOptions = Omit<UseMutationOptions<CancelRes, Error, CancelReq>, "mutationFn">;

function useCancel({ onSuccess, ...options }: CancelOptions = {}) {
    const { mutations } = useSystemTasksApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.cancel,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["operations.tasks.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["operations.tasks.find-one-by-id"]],
            });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const SystemTasksCommands = Object.freeze({
    useCancel,
});
