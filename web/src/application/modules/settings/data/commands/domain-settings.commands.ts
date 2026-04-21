import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useDomainSettingsApi } from "~/settings/api/hooks";
import type {
    DomainSettings_DeleteOne_Req,
    DomainSettings_DeleteOne_Res,
    DomainSettings_UpdateOne_Req,
    DomainSettings_UpdateOne_Res,
    DomainSettings_UpdateStatus_Req,
    DomainSettings_UpdateStatus_Res,
} from "~/settings/api/services/domain-settings-services";
import { QK } from "~/settings/data/constants";

type UpdateOneReq = DomainSettings_UpdateOne_Req["data"];
type UpdateOneRes = DomainSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useDomainSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.domain-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = DomainSettings_UpdateStatus_Req["data"];
type UpdateStatusRes = DomainSettings_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useDomainSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.domain-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = DomainSettings_DeleteOne_Req["data"];
type DeleteOneRes = DomainSettings_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useDomainSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.domain-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const DomainSettingsCommands = Object.freeze({
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});
