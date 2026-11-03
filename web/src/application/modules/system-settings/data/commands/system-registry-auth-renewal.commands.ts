import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSystemRegistryAuthRenewalApi } from "~/system-settings/api/hooks";
import type {
    SystemRegistryAuthRenewal_Execute_Req,
    SystemRegistryAuthRenewal_Execute_Res,
    SystemRegistryAuthRenewal_UpdateOne_Req,
    SystemRegistryAuthRenewal_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = SystemRegistryAuthRenewal_UpdateOne_Req["data"];
type UpdateOneRes = SystemRegistryAuthRenewal_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;
type ExecuteReq = SystemRegistryAuthRenewal_Execute_Req["data"];
type ExecuteRes = SystemRegistryAuthRenewal_Execute_Res;
type ExecuteOptions = Omit<UseMutationOptions<ExecuteRes, Error, ExecuteReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSystemRegistryAuthRenewalApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.registry-auth-renewal.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

function useExecute({ onSuccess, ...options }: ExecuteOptions = {}) {
    const { mutations } = useSystemRegistryAuthRenewalApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.execute,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.registry-auth-renewal.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SystemRegistryAuthRenewalCommands = Object.freeze({
    useUpdateOne,
    useExecute,
});
