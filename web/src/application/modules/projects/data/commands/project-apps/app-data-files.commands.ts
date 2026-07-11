import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAppDataFilesApi } from "~/projects/api/hooks/project-apps";
import type {
    AppDataFiles_CreateOne_Req,
    AppDataFiles_CreateOne_Res,
    AppDataFiles_DeleteOne_Req,
    AppDataFiles_DeleteOne_Res,
    AppDataFiles_UploadLocal_Req,
    AppDataFiles_UploadLocal_Res,
} from "~/projects/api/services";
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

type UploadLocalReq = AppDataFiles_UploadLocal_Req["data"];
type UploadLocalRes = AppDataFiles_UploadLocal_Res;
type UploadLocalOptions = Omit<UseMutationOptions<UploadLocalRes, Error, UploadLocalReq>, "mutationFn">;

function useUploadLocal({ onSuccess, ...options }: UploadLocalOptions = {}) {
    const { mutations } = useAppDataFilesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (data: UploadLocalReq) => mutations.uploadLocal(data),
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.data-files.$.find-many-paginated"]],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type CreateOneReq = AppDataFiles_CreateOne_Req["data"];
type CreateOneRes = AppDataFiles_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useAppDataFilesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: (data: CreateOneReq) => mutations.createOne(data),
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
    useUploadLocal,
    useCreateOne,
});
