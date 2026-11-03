import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectGithubAppApi } from "~/projects/api/hooks";
import type {
    ProjectGithubApp_BeginManifestFlow_Req,
    ProjectGithubApp_BeginManifestFlow_Res,
    ProjectGithubApp_BeginReprovision_Req,
    ProjectGithubApp_BeginReprovision_Res,
    ProjectGithubApp_CreateOne_Req,
    ProjectGithubApp_CreateOne_Res,
    ProjectGithubApp_DeleteOne_Req,
    ProjectGithubApp_DeleteOne_Res,
    ProjectGithubApp_UpdateOne_Req,
    ProjectGithubApp_UpdateOne_Res,
    ProjectGithubApp_UpdateStatus_Req,
    ProjectGithubApp_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type CreateOneReq = ProjectGithubApp_CreateOne_Req["data"];
type CreateOneRes = ProjectGithubApp_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectGithubAppApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.github-app.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = ProjectGithubApp_UpdateOne_Req["data"];
type UpdateOneRes = ProjectGithubApp_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectGithubAppApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.github-app.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.github-app.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = ProjectGithubApp_UpdateStatus_Req["data"];
type UpdateStatusRes = ProjectGithubApp_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useProjectGithubAppApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.github-app.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.github-app.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ProjectGithubApp_DeleteOne_Req["data"];
type DeleteOneRes = ProjectGithubApp_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectGithubAppApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.github-app.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.github-app.$.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type BeginManifestFlowReq = ProjectGithubApp_BeginManifestFlow_Req["data"];
type BeginManifestFlowRes = ProjectGithubApp_BeginManifestFlow_Res;
type BeginManifestFlowOptions = Omit<
    UseMutationOptions<BeginManifestFlowRes, Error, BeginManifestFlowReq>,
    "mutationFn"
>;

function useBeginManifestFlow(options: BeginManifestFlowOptions = {}) {
    const { mutations } = useProjectGithubAppApi();

    return useMutation({
        mutationFn: mutations.beginManifestFlow,
        ...options,
    });
}

type BeginReprovisionReq = ProjectGithubApp_BeginReprovision_Req["data"];
type BeginReprovisionRes = ProjectGithubApp_BeginReprovision_Res;
type BeginReprovisionOptions = Omit<UseMutationOptions<BeginReprovisionRes, Error, BeginReprovisionReq>, "mutationFn">;

function useBeginReprovision(options: BeginReprovisionOptions = {}) {
    const { mutations } = useProjectGithubAppApi();

    return useMutation({
        mutationFn: mutations.beginReprovision,
        ...options,
    });
}

export const ProjectGithubAppCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
    useBeginManifestFlow,
    useBeginReprovision,
});
