import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";

import { useSessionApi } from "@application/shared/api";
import { type Session_Logout_Res } from "@application/shared/api/services";

import { isSessionInvalidException } from "@infrastructure/api";

/**
 * Logout command
 */
type LogoutRes = Session_Logout_Res;

type LogoutOptions = Omit<UseMutationOptions<LogoutRes>, "mutationFn"> & {
    onSessionInvalid?: () => void;
};

function useLogout({ onSuccess, onError, onSessionInvalid, ...options }: LogoutOptions = {}) {
    const queryClient = useQueryClient();

    const {
        mutations: { logout },
    } = useSessionApi();

    return useMutation({
        mutationFn: logout,
        onSuccess: (...rest) => {
            queryClient.clear();

            if (onSuccess) onSuccess(...rest);
        },
        onError: (error, ...rest) => {
            if (isSessionInvalidException(error)) {
                queryClient.clear();

                if (onSessionInvalid) onSessionInvalid();
            }

            if (onError) onError(error, ...rest);
        },

        ...options,
    });
}

export const SessionCommands = Object.freeze({
    useLogout,
});
