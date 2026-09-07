import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useHivePaaSRoutingSettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSRoutingSettings_ConfirmChange_Req,
    HivePaaSRoutingSettings_ConfirmChange_Res,
    HivePaaSRoutingSettings_RevertChange_Req,
    HivePaaSRoutingSettings_RevertChange_Res,
    HivePaaSRoutingSettings_UpdateOne_Req,
    HivePaaSRoutingSettings_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = HivePaaSRoutingSettings_UpdateOne_Req["data"];
type UpdateOneRes = HivePaaSRoutingSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useHivePaaSRoutingSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.routing-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type ConfirmChangeReq = HivePaaSRoutingSettings_ConfirmChange_Req["data"];
type ConfirmChangeRes = HivePaaSRoutingSettings_ConfirmChange_Res;
type ConfirmChangeOptions = Omit<UseMutationOptions<ConfirmChangeRes, Error, ConfirmChangeReq>, "mutationFn">;

function useConfirmChange({ onSuccess, ...options }: ConfirmChangeOptions = {}) {
    const { mutations } = useHivePaaSRoutingSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.confirmChange,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.routing-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        // Retries are the dialog's job here, not the query client's. It knows when
        // the change becomes confirmable and how long is left before the deadline;
        // a blind retry would spend that budget on requests that cannot succeed yet.
        retry: false,
        ...options,
    });
}

type RevertChangeReq = HivePaaSRoutingSettings_RevertChange_Req["data"];
type RevertChangeRes = HivePaaSRoutingSettings_RevertChange_Res;
type RevertChangeOptions = Omit<UseMutationOptions<RevertChangeRes, Error, RevertChangeReq>, "mutationFn">;

function useRevertChange({ onSuccess, ...options }: RevertChangeOptions = {}) {
    const { mutations } = useHivePaaSRoutingSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.revertChange,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.hivepaas.routing-settings.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        retry: false,
        ...options,
    });
}

export const HivePaaSRoutingSettingsCommands = Object.freeze({
    useUpdateOne,
    useConfirmChange,
    useRevertChange,
});

export { HivePaaSRoutingSettingsCommands as HivePaaSHttpSettingsCommands };
