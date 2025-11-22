import { type UseMutationOptions, useMutation } from "@tanstack/react-query";

import { useProfileApi, useSessionApi } from "@application/shared/api";
import {
    type Profile_Complete2FASetup_Req,
    type Profile_Complete2FASetup_Res,
    type Profile_GetProfile2FASetup_Req,
    type Profile_GetProfile2FASetup_Res,
    type Profile_UpdateProfile_Req,
    type Profile_UpdateProfile_Res,
} from "@application/shared/api/services";
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

export const ProfileCommands = Object.freeze({
    useGetProfile2FASetup,
    useComplete2FASetup,
    useUpdate,
});
