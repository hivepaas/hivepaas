import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useHivePaaSRegistrySettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSRegistrySettings_CheckPush_Req,
    HivePaaSRegistrySettings_CheckPush_Res,
    HivePaaSRegistrySettings_ProbeDomain_Req,
    HivePaaSRegistrySettings_ProbeDomain_Res,
    HivePaaSRegistrySettings_RotateCredential_Req,
    HivePaaSRegistrySettings_RotateCredential_Res,
    HivePaaSRegistrySettings_UpdateOne_Req,
    HivePaaSRegistrySettings_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = HivePaaSRegistrySettings_UpdateOne_Req["data"];
type UpdateOneRes = HivePaaSRegistrySettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useHivePaaSRegistrySettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.hivepaas.registry.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type ProbeDomainReq = HivePaaSRegistrySettings_ProbeDomain_Req["data"];
type ProbeDomainRes = HivePaaSRegistrySettings_ProbeDomain_Res;

function useProbeDomain(options: Omit<UseMutationOptions<ProbeDomainRes, Error, ProbeDomainReq>, "mutationFn"> = {}) {
    const { mutations } = useHivePaaSRegistrySettingsApi();

    return useMutation({ mutationFn: mutations.probeDomain, ...options });
}

type CheckPushReq = HivePaaSRegistrySettings_CheckPush_Req["data"];
type CheckPushRes = HivePaaSRegistrySettings_CheckPush_Res;

function useCheckPush(options: Omit<UseMutationOptions<CheckPushRes, Error, CheckPushReq>, "mutationFn"> = {}) {
    const { mutations } = useHivePaaSRegistrySettingsApi();

    return useMutation({ mutationFn: mutations.checkPush, ...options });
}

type RotateCredentialReq = HivePaaSRegistrySettings_RotateCredential_Req["data"];
type RotateCredentialRes = HivePaaSRegistrySettings_RotateCredential_Res;

function useRotateCredential({
    onSuccess,
    ...options
}: Omit<UseMutationOptions<RotateCredentialRes, Error, RotateCredentialReq>, "mutationFn"> = {}) {
    const { mutations } = useHivePaaSRegistrySettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.rotateCredential,
        onSuccess: (response, ...rest) => {
            // The page shows when the password last changed and how long the old
            // one still works, so it has to be read again.
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.hivepaas.registry.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const HivePaaSRegistrySettingsCommands = Object.freeze({
    useUpdateOne,
    useProbeDomain,
    useCheckPush,
    useRotateCredential,
});
