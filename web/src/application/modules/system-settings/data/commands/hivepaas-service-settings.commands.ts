import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useHivePaaSServiceSettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSServiceSettings_ConfirmChange_Req,
    HivePaaSServiceSettings_ConfirmChange_Res,
    HivePaaSServiceSettings_RevertChange_Req,
    HivePaaSServiceSettings_RevertChange_Res,
    HivePaaSServiceSettings_UpdateOne_Req,
    HivePaaSServiceSettings_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = HivePaaSServiceSettings_UpdateOne_Req["data"];
type UpdateOneRes = HivePaaSServiceSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useHivePaaSServiceSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.service-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type ConfirmChangeReq = HivePaaSServiceSettings_ConfirmChange_Req["data"];
type ConfirmChangeRes = HivePaaSServiceSettings_ConfirmChange_Res;
type ConfirmChangeOptions = Omit<UseMutationOptions<ConfirmChangeRes, Error, ConfirmChangeReq>, "mutationFn">;

function useConfirmChange({ onSuccess, ...options }: ConfirmChangeOptions = {}) {
    const { mutations } = useHivePaaSServiceSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.confirmChange,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.service-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        // Retries belong to the dialog: it knows when the change becomes
        // confirmable and how long is left, and a blind retry would spend that
        // budget on requests that cannot succeed yet.
        retry: false,
        ...options,
    });
}

type RevertChangeReq = HivePaaSServiceSettings_RevertChange_Req["data"];
type RevertChangeRes = HivePaaSServiceSettings_RevertChange_Res;
type RevertChangeOptions = Omit<UseMutationOptions<RevertChangeRes, Error, RevertChangeReq>, "mutationFn">;

function useRevertChange({ onSuccess, ...options }: RevertChangeOptions = {}) {
    const { mutations } = useHivePaaSServiceSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.revertChange,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.service-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        retry: false,
        ...options,
    });
}

export const HivePaaSServiceSettingsCommands = Object.freeze({
    useUpdateOne,
    useConfirmChange,
    useRevertChange,
});
