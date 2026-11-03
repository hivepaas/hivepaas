import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useLicenseApi } from "@application/shared/api";
import type {
    License_InstallKey_Req,
    License_InstallKey_Res,
    License_Refresh_Res,
} from "@application/shared/api/services";
import { QK } from "@application/shared/data/constants";

/*
 * Each command invalidates the one license key and nothing else. System ›
 * License, the banners, the plan-limit dialog and the OAuth page all read it, so
 * that one invalidation updates every one of them.
 */

/**
 * Install a license key
 */
type InstallKeyReq = License_InstallKey_Req["data"];
type InstallKeyRes = License_InstallKey_Res;

type InstallKeyOptions = Omit<UseMutationOptions<InstallKeyRes, Error, InstallKeyReq>, "mutationFn">;

function useInstallKey({ onSuccess, ...options }: InstallKeyOptions = {}) {
    const { mutations } = useLicenseApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.installKey,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["license.find-one"]] });

            onSuccess?.(response, ...rest);
        },

        ...options,
    });
}

/**
 * Ask the license portal for a newer key
 */
type RefreshRes = License_Refresh_Res;

type RefreshOptions = Omit<UseMutationOptions<RefreshRes>, "mutationFn">;

function useRefresh({ onSuccess, onError, ...options }: RefreshOptions = {}) {
    const { mutations } = useLicenseApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.refresh,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["license.find-one"]] });

            onSuccess?.(response, ...rest);
        },
        // A check that failed is recorded as well - when it was made, and why it
        // failed - and the page shows both.
        onError: (error, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["license.find-one"]] });

            onError?.(error, ...rest);
        },

        ...options,
    });
}

export const LicenseCommands = Object.freeze({
    useInstallKey,
    useRefresh,
});
