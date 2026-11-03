import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectSslProviderApi } from "~/projects/api/hooks";
import type {
    ProjectSslProvider_CreateOne_Req,
    ProjectSslProvider_CreateOne_Res,
    ProjectSslProvider_DeleteOne_Req,
    ProjectSslProvider_DeleteOne_Res,
    ProjectSslProvider_UpdateOne_Req,
    ProjectSslProvider_UpdateOne_Res,
    ProjectSslProvider_UpdateStatus_Req,
    ProjectSslProvider_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type CreateOneReq = ProjectSslProvider_CreateOne_Req["data"];
type CreateOneRes = ProjectSslProvider_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectSslProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.ssl-provider.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = ProjectSslProvider_UpdateOne_Req["data"];
type UpdateOneRes = ProjectSslProvider_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectSslProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.ssl-provider.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.ssl-provider.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = ProjectSslProvider_UpdateStatus_Req["data"];
type UpdateStatusRes = ProjectSslProvider_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useProjectSslProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.ssl-provider.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.ssl-provider.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ProjectSslProvider_DeleteOne_Req["data"];
type DeleteOneRes = ProjectSslProvider_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectSslProviderApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.ssl-provider.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.ssl-provider.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const ProjectSslProviderCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});
