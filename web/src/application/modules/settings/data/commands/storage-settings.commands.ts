import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useStorageSettingsApi } from "~/settings/api/hooks";
import type {
    StorageSettings_DeleteOne_Req,
    StorageSettings_DeleteOne_Res,
    StorageSettings_UpdateOne_Req,
    StorageSettings_UpdateOne_Res,
    StorageSettings_UpdateStatus_Req,
    StorageSettings_UpdateStatus_Res,
} from "~/settings/api/services/storage-settings-services";
import { QK } from "~/settings/data/constants";

type UpdateOneReq = StorageSettings_UpdateOne_Req["data"];
type UpdateOneRes = StorageSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useStorageSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.storage-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = StorageSettings_UpdateStatus_Req["data"];
type UpdateStatusRes = StorageSettings_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useStorageSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.storage-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = StorageSettings_DeleteOne_Req["data"];
type DeleteOneRes = StorageSettings_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useStorageSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.storage-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const StorageSettingsCommands = Object.freeze({
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});
