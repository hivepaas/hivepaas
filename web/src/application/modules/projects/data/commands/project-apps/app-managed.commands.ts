import { type QueryClient, type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useAppManagedApi } from "../../../api/hooks/project-apps";
import type {
    AppManaged_CreateRestorePoint_Req,
    AppManaged_CreateRestorePoint_Res,
    AppManaged_DeleteRestorePoint_Req,
    AppManaged_DeleteRestorePoint_Res,
    AppManaged_RefreshCredentials_Req,
    AppManaged_RefreshCredentials_Res,
    AppManaged_Restore_Req,
    AppManaged_Restore_Res,
    AppManaged_Switch_Req,
    AppManaged_Switch_Res,
    AppManaged_Unmanage_Req,
    AppManaged_Unmanage_Res,
    AppManaged_UpdateRestorePoint_Req,
    AppManaged_UpdateRestorePoint_Res,
    AppManaged_UpdateSettings_Req,
    AppManaged_UpdateSettings_Res,
} from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

import { invalidateSingleAppConfigurationQueries } from "./app-configuration-cache.helpers";

type Options<Res, Req> = Omit<UseMutationOptions<Res, Error, Req>, "mutationFn">;
type AppScope = { projectID: string; appID: string };

function invalidateManaged(queryClient: QueryClient, scope: AppScope): void {
    void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.managed.$.find-settings"], scope] });
    void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.managed.$.find-status"], scope] });
}

function invalidateManagedJobs(queryClient: QueryClient, scope: AppScope): void {
    void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.scheduled-jobs.$.find-many-paginated"]] });
    invalidateManaged(queryClient, scope);
}

function useUpdateSettings({
    onSuccess,
    ...options
}: Options<AppManaged_UpdateSettings_Res, AppManaged_UpdateSettings_Req["data"]> = {}) {
    const { mutations } = useAppManagedApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateSettings,
        onSuccess: (response, request, ...rest) => {
            // A move of the restore test makes or deletes a verifier app.
            void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.$.find-many-paginated"]] });
            invalidateManagedJobs(queryClient, request);
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

function useRefreshCredentials({
    onSuccess,
    ...options
}: Options<AppManaged_RefreshCredentials_Res, AppManaged_RefreshCredentials_Req["data"]> = {}) {
    const { mutations } = useAppManagedApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.refreshCredentials,
        onSuccess: (response, request, ...rest) => {
            invalidateSingleAppConfigurationQueries(queryClient, request);
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

function useUnmanage({
    onSuccess,
    ...options
}: Options<AppManaged_Unmanage_Res, AppManaged_Unmanage_Req["data"]> = {}) {
    const { mutations } = useAppManagedApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.unmanage,
        onSuccess: (response, request, ...rest) => {
            invalidateManagedJobs(queryClient, request);
            invalidateSingleAppConfigurationQueries(queryClient, request);
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

function useRestore({ onSuccess, ...options }: Options<AppManaged_Restore_Res, AppManaged_Restore_Req["data"]> = {}) {
    const { mutations } = useAppManagedApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.restore,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.$.find-many-paginated"]] });
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

function useSwitch({ onSuccess, ...options }: Options<AppManaged_Switch_Res, AppManaged_Switch_Req["data"]> = {}) {
    const { mutations } = useAppManagedApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.switchConsumers,
        onSuccess: (response, request, ...rest) => {
            // Both databases' settings say so, the source's jobs pause, and the
            // consumers' variables and settings change.
            void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.managed.$.find-settings"]] });
            void queryClient.invalidateQueries({ queryKey: [QK["projects.apps.managed.$.find-switch"]] });
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.apps.scheduled-jobs.$.find-many-paginated"]],
            });
            invalidateManaged(queryClient, request);
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

function useCreateRestorePoint({
    onSuccess,
    ...options
}: Options<AppManaged_CreateRestorePoint_Res, AppManaged_CreateRestorePoint_Req["data"]> = {}) {
    const { mutations } = useAppManagedApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createRestorePoint,
        onSuccess: (response, request, ...rest) => {
            invalidateManaged(queryClient, request);
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

function useUpdateRestorePoint({
    onSuccess,
    ...options
}: Options<AppManaged_UpdateRestorePoint_Res, AppManaged_UpdateRestorePoint_Req["data"]> = {}) {
    const { mutations } = useAppManagedApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateRestorePoint,
        onSuccess: (response, request, ...rest) => {
            invalidateManaged(queryClient, request);
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

function useDeleteRestorePoint({
    onSuccess,
    ...options
}: Options<AppManaged_DeleteRestorePoint_Res, AppManaged_DeleteRestorePoint_Req["data"]> = {}) {
    const { mutations } = useAppManagedApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteRestorePoint,
        onSuccess: (response, request, ...rest) => {
            invalidateManaged(queryClient, request);
            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const AppManagedCommands = Object.freeze({
    useSwitch,
    useUpdateSettings,
    useRefreshCredentials,
    useUnmanage,
    useRestore,
    useCreateRestorePoint,
    useUpdateRestorePoint,
    useDeleteRestorePoint,
});
