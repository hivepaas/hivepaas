import { type UseMutationOptions, useMutation } from "@tanstack/react-query";

import { useProfileApi } from "@application/shared/api";
import { type Profile_GetProfile2FASetup_Res } from "@application/shared/api/services";

/**
 * Update profile
 */
type GetProfile2FASetupRes = Profile_GetProfile2FASetup_Res;

type GetProfile2FASetupOptions = Omit<UseMutationOptions<GetProfile2FASetupRes, Error, undefined>, "mutationFn">;

function useGetProfile2FASetup({ onSuccess, ...options }: GetProfile2FASetupOptions = {}) {
    const {
        mutations: { getProfile2FASetup },
    } = useProfileApi();

    return useMutation({
        mutationFn: getProfile2FASetup,
        onSuccess,
        ...options,
    });
}

export const ProfileCommands = Object.freeze({
    useGetProfile2FASetup,
});
