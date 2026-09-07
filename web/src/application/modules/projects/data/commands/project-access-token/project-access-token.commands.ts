import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectAccessTokenApi } from "~/projects/api/hooks";
import type {
    ProjectAccessToken_CreateOne_Req,
    ProjectAccessToken_CreateOne_Res,
    ProjectAccessToken_DeleteOne_Req,
    ProjectAccessToken_DeleteOne_Res,
    ProjectAccessToken_UpdateOne_Req,
    ProjectAccessToken_UpdateOne_Res,
    ProjectAccessToken_UpdateStatus_Req,
    ProjectAccessToken_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type CreateOneReq = ProjectAccessToken_CreateOne_Req["data"];
type CreateOneRes = ProjectAccessToken_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectAccessTokenApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.access-token.$.find-many-paginated"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = ProjectAccessToken_UpdateOne_Req["data"];
type UpdateOneRes = ProjectAccessToken_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectAccessTokenApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.access-token.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.access-token.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = ProjectAccessToken_UpdateStatus_Req["data"];
type UpdateStatusRes = ProjectAccessToken_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useProjectAccessTokenApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.access-token.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.access-token.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ProjectAccessToken_DeleteOne_Req["data"];
type DeleteOneRes = ProjectAccessToken_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectAccessTokenApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.access-token.$.find-many-paginated"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.access-token.$.find-one-by-id"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const ProjectAccessTokenCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});
