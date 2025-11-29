import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useProfileApi, useSessionApi } from "@application/shared/api";
import type {
    Profile_Complete2FASetup_Req,
    Profile_Complete2FASetup_Res,
    Profile_CreateOneApiKey_Req,
    Profile_CreateOneApiKey_Res,
    Profile_DeleteOneApiKey_Req,
    Profile_DeleteOneApiKey_Res,
    Profile_GetProfile2FASetup_Req,
    Profile_GetProfile2FASetup_Res,
    Profile_UpdateOneApiKeyStatus_Req,
    Profile_UpdateOneApiKeyStatus_Res,
    Profile_UpdateProfilePassword_Req,
    Profile_UpdateProfilePassword_Res,
    Profile_UpdateProfile_Req,
    Profile_UpdateProfile_Res,
} from "@application/shared/api/services";
import { QK } from "@application/shared/data/constants";
import { type Profile } from "@application/shared/entities";

/**
 * Update profile
 */
type GetProfile2FASetupReq = Profile_GetProfile2FASetup_Req["data"];
type GetProfile2FASetupRes = Profile_GetProfile2FASetup_Res;

type GetProfile2FASetupOptions = Omit<
    UseMutationOptions<GetProfile2FASetupRes, Error, GetProfile2FASetupReq>,
    "mutationFn"
>;

function useGetProfile2FASetup({ onSuccess, ...options }: GetProfile2FASetupOptions = {}) {
    const {
        mutations: { getProfile2FASetup },
    } = useProfileApi();

    return useMutation({
        mutationFn: getProfile2FASetup,
        onSuccess: (response, request, ...rest) => {
            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },
        ...options,
    });
}

/**
 * Update profile
 */
type Complete2FASetupReq = Profile_Complete2FASetup_Req["data"];
type Complete2FASetupRes = Profile_Complete2FASetup_Res;

type Complete2FASetupOptions = Omit<UseMutationOptions<Complete2FASetupRes, Error, Complete2FASetupReq>, "mutationFn">;

function useComplete2FASetup({ onSuccess, ...options }: Complete2FASetupOptions = {}) {
    const {
        mutations: { complete2FASetup },
    } = useProfileApi();

    return useMutation({
        mutationFn: complete2FASetup,
        onSuccess: (response, request, ...rest) => {
            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },

        ...options,
    });
}

/**
 * Update profile command
 */
type UpdateProfileReq = Profile_UpdateProfile_Req["data"];
type UpdateProfileRes = Profile_UpdateProfile_Res;
type UpdateProfileOptions = Omit<
    UseMutationOptions<UpdateProfileRes, Error, UpdateProfileReq>,
    "mutationFn" | "onSuccess"
> & {
    onSuccess?: (profile: Profile) => void;
};

function useUpdate({ onSuccess, ...options }: UpdateProfileOptions = {}) {
    const { mutations } = useProfileApi();

    const {
        queries: { getProfile },
    } = useSessionApi();

    async function updateProfileFn(values: UpdateProfileReq) {
        const res = await mutations.update(values);

        const { data: profile } = await getProfile();

        if (onSuccess) onSuccess(profile);

        return res;
    }

    return useMutation({
        mutationFn: updateProfileFn,

        ...options,
    });
}

/**
 * Update profile password
 */
type UpdatePasswordReq = Profile_UpdateProfilePassword_Req["data"];
type UpdatePasswordRes = Profile_UpdateProfilePassword_Res;

type UpdatePasswordOptions = Omit<UseMutationOptions<UpdatePasswordRes, Error, UpdatePasswordReq>, "mutationFn">;

function useUpdatePassword(options: UpdatePasswordOptions = {}) {
    const {
        mutations: { updatePassword },
    } = useProfileApi();

    return useMutation({
        mutationFn: updatePassword,

        ...options,
    });
}

/**
 * Create one profile API key
 */
type CreateOneApiKeyReq = Profile_CreateOneApiKey_Req["data"];
type CreateOneApiKeyRes = Profile_CreateOneApiKey_Res;

type CreateOneApiKeyOptions = Omit<UseMutationOptions<CreateOneApiKeyRes, Error, CreateOneApiKeyReq>, "mutationFn">;

function useCreateOneApiKey({ onSuccess, ...options }: CreateOneApiKeyOptions = {}) {
    const { mutations } = useProfileApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOneApiKey,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["profile-api-keys.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }

            return response;
        },

        ...options,
    });
}

/**
 * Delete one profile API key
 */
type DeleteOneApiKeyReq = Profile_DeleteOneApiKey_Req["data"];
type DeleteOneApiKeyRes = Profile_DeleteOneApiKey_Res;

type DeleteOneApiKeyOptions = Omit<UseMutationOptions<DeleteOneApiKeyRes, Error, DeleteOneApiKeyReq>, "mutationFn">;

function useDeleteOneApiKey({ onSuccess, ...options }: DeleteOneApiKeyOptions = {}) {
    const { mutations } = useProfileApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOneApiKey,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["profile-api-keys.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },

        ...options,
    });
}

/**
 * Update one profile API key status
 */
type UpdateOneApiKeyStatusReq = Profile_UpdateOneApiKeyStatus_Req["data"];
type UpdateOneApiKeyStatusRes = Profile_UpdateOneApiKeyStatus_Res;

type UpdateOneApiKeyStatusOptions = Omit<
    UseMutationOptions<UpdateOneApiKeyStatusRes, Error, UpdateOneApiKeyStatusReq>,
    "mutationFn"
>;

function useUpdateOneApiKeyStatus({ onSuccess, ...options }: UpdateOneApiKeyStatusOptions = {}) {
    const { mutations } = useProfileApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOneApiKeyStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["profile-api-keys.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },

        ...options,
    });
}

export const ProfileCommands = Object.freeze({
    useGetProfile2FASetup,
    useComplete2FASetup,
    useUpdate,
    useUpdatePassword,
    useCreateOneApiKey,
    useDeleteOneApiKey,
    useUpdateOneApiKeyStatus,
});
