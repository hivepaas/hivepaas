import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectRegistryAuthApi } from "~/projects/api/hooks";
import type {
    ProjectRegistryAuth_CreateOne_Req,
    ProjectRegistryAuth_CreateOne_Res,
    ProjectRegistryAuth_DeleteOne_Req,
    ProjectRegistryAuth_DeleteOne_Res,
    ProjectRegistryAuth_UpdateMeta_Req,
    ProjectRegistryAuth_UpdateMeta_Res,
    ProjectRegistryAuth_UpdateOne_Req,
    ProjectRegistryAuth_UpdateOne_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type CreateOneReq = ProjectRegistryAuth_CreateOne_Req["data"];
type CreateOneRes = ProjectRegistryAuth_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectRegistryAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.registry-auth.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = ProjectRegistryAuth_UpdateOne_Req["data"];
type UpdateOneRes = ProjectRegistryAuth_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectRegistryAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.registry-auth.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.registry-auth.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateMetaReq = ProjectRegistryAuth_UpdateMeta_Req["data"];
type UpdateMetaRes = ProjectRegistryAuth_UpdateMeta_Res;
type UpdateMetaOptions = Omit<UseMutationOptions<UpdateMetaRes, Error, UpdateMetaReq>, "mutationFn">;

function useUpdateMeta({ onSuccess, ...options }: UpdateMetaOptions = {}) {
    const { mutations } = useProjectRegistryAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateMeta,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.registry-auth.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.registry-auth.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ProjectRegistryAuth_DeleteOne_Req["data"];
type DeleteOneRes = ProjectRegistryAuth_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectRegistryAuthApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.registry-auth.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.registry-auth.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const ProjectRegistryAuthCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateMeta,
    useDeleteOne,
});
