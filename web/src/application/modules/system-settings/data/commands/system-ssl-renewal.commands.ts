import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSystemSslRenewalApi } from "~/system-settings/api/hooks";
import type {
    SystemSslRenewal_Execute_Req,
    SystemSslRenewal_Execute_Res,
    SystemSslRenewal_UpdateOne_Req,
    SystemSslRenewal_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = SystemSslRenewal_UpdateOne_Req["data"];
type UpdateOneRes = SystemSslRenewal_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;
type ExecuteReq = SystemSslRenewal_Execute_Req["data"];
type ExecuteRes = SystemSslRenewal_Execute_Res;
type ExecuteOptions = Omit<UseMutationOptions<ExecuteRes, Error, ExecuteReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSystemSslRenewalApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.ssl-renewal.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

function useExecute({ onSuccess, ...options }: ExecuteOptions = {}) {
    const { mutations } = useSystemSslRenewalApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.execute,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.ssl-renewal.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SystemSslRenewalCommands = Object.freeze({
    useUpdateOne,
    useExecute,
});
