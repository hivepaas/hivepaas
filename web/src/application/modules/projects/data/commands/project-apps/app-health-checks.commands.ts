import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAppHealthChecksApi } from "~/projects/api/hooks/project-apps";
import type {
    AppHealthChecks_CreateOne_Req,
    AppHealthChecks_CreateOne_Res,
    AppHealthChecks_DeleteOne_Req,
    AppHealthChecks_DeleteOne_Res,
    AppHealthChecks_UpdateOne_Req,
    AppHealthChecks_UpdateOne_Res,
    AppHealthChecks_UpdateStatus_Req,
    AppHealthChecks_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type CreateOneReq = AppHealthChecks_CreateOne_Req["data"];
type CreateOneRes = AppHealthChecks_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useAppHealthChecksApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.health-checks.$.find-many-paginated"]],
            });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = AppHealthChecks_UpdateOne_Req["data"];
type UpdateOneRes = AppHealthChecks_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useAppHealthChecksApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.health-checks.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.health-checks.$.find-one-by-id"]],
            });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = AppHealthChecks_UpdateStatus_Req["data"];
type UpdateStatusRes = AppHealthChecks_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useAppHealthChecksApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.health-checks.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.health-checks.$.find-one-by-id"]],
            });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = AppHealthChecks_DeleteOne_Req["data"];
type DeleteOneRes = AppHealthChecks_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useAppHealthChecksApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.health-checks.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.health-checks.$.find-one-by-id"]],
            });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const AppHealthChecksCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});
