import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectDomainSettingsApi } from "~/projects/api/hooks";
import type {
    ProjectDomainSettings_DeleteOne_Req,
    ProjectDomainSettings_DeleteOne_Res,
    ProjectDomainSettings_UpdateOne_Req,
    ProjectDomainSettings_UpdateOne_Res,
    ProjectDomainSettings_UpdateStatus_Req,
    ProjectDomainSettings_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type UpdateOneReq = ProjectDomainSettings_UpdateOne_Req["data"];
type UpdateOneRes = ProjectDomainSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectDomainSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.domain-settings.$.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = ProjectDomainSettings_UpdateStatus_Req["data"];
type UpdateStatusRes = ProjectDomainSettings_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useProjectDomainSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.domain-settings.$.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ProjectDomainSettings_DeleteOne_Req["data"];
type DeleteOneRes = ProjectDomainSettings_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectDomainSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.domain-settings.$.find-one"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const ProjectDomainSettingsCommands = Object.freeze({
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});
