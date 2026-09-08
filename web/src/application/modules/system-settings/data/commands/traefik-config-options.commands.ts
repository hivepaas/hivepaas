import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useTraefikConfigOptionsApi } from "~/system-settings/api/hooks";
import type {
    TraefikConfigOptions_ConfirmChange_Req,
    TraefikConfigOptions_ConfirmChange_Res,
    TraefikConfigOptions_RevertChange_Req,
    TraefikConfigOptions_RevertChange_Res,
    TraefikConfigOptions_UpdateOne_Req,
    TraefikConfigOptions_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = TraefikConfigOptions_UpdateOne_Req["data"];
type UpdateOneRes = TraefikConfigOptions_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useTraefikConfigOptionsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.traefik.config-options.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type ConfirmChangeReq = TraefikConfigOptions_ConfirmChange_Req["data"];
type ConfirmChangeRes = TraefikConfigOptions_ConfirmChange_Res;
type ConfirmChangeOptions = Omit<UseMutationOptions<ConfirmChangeRes, Error, ConfirmChangeReq>, "mutationFn">;

function useConfirmChange({ onSuccess, ...options }: ConfirmChangeOptions = {}) {
    const { mutations } = useTraefikConfigOptionsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.confirmChange,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.traefik.config-options.find-one"]],
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

type RevertChangeReq = TraefikConfigOptions_RevertChange_Req["data"];
type RevertChangeRes = TraefikConfigOptions_RevertChange_Res;
type RevertChangeOptions = Omit<UseMutationOptions<RevertChangeRes, Error, RevertChangeReq>, "mutationFn">;

function useRevertChange({ onSuccess, ...options }: RevertChangeOptions = {}) {
    const { mutations } = useTraefikConfigOptionsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.revertChange,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["system-settings.traefik.config-options.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        retry: false,
        ...options,
    });
}

export const TraefikConfigOptionsCommands = Object.freeze({
    useUpdateOne,
    useConfirmChange,
    useRevertChange,
});
