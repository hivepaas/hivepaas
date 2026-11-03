import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAppPreviewsApi } from "~/projects/api/hooks/project-apps";
import type {
    AppPreviews_CreateOne_Req,
    AppPreviews_CreateOne_Res,
    AppPreviews_PrepareCreate_Req,
    AppPreviews_PrepareCreate_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type PrepareCreateReq = AppPreviews_PrepareCreate_Req["data"];
type PrepareCreateRes = AppPreviews_PrepareCreate_Res;
type PrepareCreateOptions = Omit<UseMutationOptions<PrepareCreateRes, Error, PrepareCreateReq>, "mutationFn">;

function usePrepareCreate({ onSuccess, ...options }: PrepareCreateOptions = {}) {
    const { mutations } = useAppPreviewsApi();

    return useMutation({
        mutationFn: mutations.prepareCreate,
        onSuccess: (response, request, ...rest) => {
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type CreateOneReq = AppPreviews_CreateOne_Req["data"];
type CreateOneRes = AppPreviews_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useAppPreviewsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.previews.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.$.find-many-paginated"]],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppPreviewsCommands = Object.freeze({
    usePrepareCreate,
    useCreateOne,
});
