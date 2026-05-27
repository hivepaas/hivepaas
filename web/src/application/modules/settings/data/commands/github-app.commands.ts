import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { QK as PROJECT_QK } from "~/projects/data/constants";
import { useGithubAppApi } from "~/settings/api/hooks";
import type {
    GithubApp_BeginManifestFlow_Req,
    GithubApp_BeginManifestFlow_Res,
    GithubApp_BeginReprovision_Req,
    GithubApp_BeginReprovision_Res,
    GithubApp_CreateOne_Req,
    GithubApp_CreateOne_Res,
    GithubApp_DeleteOne_Req,
    GithubApp_DeleteOne_Res,
    GithubApp_TestConnection_Req,
    GithubApp_TestConnection_Res,
    GithubApp_UpdateOne_Req,
    GithubApp_UpdateOne_Res,
    GithubApp_UpdateStatus_Req,
    GithubApp_UpdateStatus_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = GithubApp_CreateOne_Req["data"];
type CreateOneRes = GithubApp_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useGithubAppApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.github-app.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [PROJECT_QK["projects.github-app.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = GithubApp_UpdateOne_Req["data"];
type UpdateOneRes = GithubApp_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useGithubAppApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.github-app.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.github-app.find-one-by-id"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [PROJECT_QK["projects.github-app.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = GithubApp_UpdateStatus_Req["data"];
type UpdateStatusRes = GithubApp_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useGithubAppApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.github-app.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.github-app.find-one-by-id"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [PROJECT_QK["projects.github-app.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = GithubApp_DeleteOne_Req["data"];
type DeleteOneRes = GithubApp_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useGithubAppApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.github-app.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.github-app.find-one-by-id"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [PROJECT_QK["projects.github-app.$.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type TestConnectionReq = GithubApp_TestConnection_Req["data"];
type TestConnectionRes = GithubApp_TestConnection_Res;
type TestConnectionOptions = Omit<UseMutationOptions<TestConnectionRes, Error, TestConnectionReq>, "mutationFn">;

function useTestConnection(options: TestConnectionOptions = {}) {
    const { mutations } = useGithubAppApi();

    return useMutation({
        mutationFn: mutations.testConnection,
        ...options,
    });
}

type BeginManifestFlowReq = GithubApp_BeginManifestFlow_Req["data"];
type BeginManifestFlowRes = GithubApp_BeginManifestFlow_Res;
type BeginManifestFlowOptions = Omit<
    UseMutationOptions<BeginManifestFlowRes, Error, BeginManifestFlowReq>,
    "mutationFn"
>;

function useBeginManifestFlow(options: BeginManifestFlowOptions = {}) {
    const { mutations } = useGithubAppApi();

    return useMutation({
        mutationFn: mutations.beginManifestFlow,
        ...options,
    });
}

type BeginReprovisionReq = GithubApp_BeginReprovision_Req["data"];
type BeginReprovisionRes = GithubApp_BeginReprovision_Res;
type BeginReprovisionOptions = Omit<UseMutationOptions<BeginReprovisionRes, Error, BeginReprovisionReq>, "mutationFn">;

function useBeginReprovision(options: BeginReprovisionOptions = {}) {
    const { mutations } = useGithubAppApi();

    return useMutation({
        mutationFn: mutations.beginReprovision,
        ...options,
    });
}

export const GithubAppCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
    useTestConnection,
    useBeginManifestFlow,
    useBeginReprovision,
});
