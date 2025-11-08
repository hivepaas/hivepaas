import { type QueryFunctionContext, type UseQueryOptions, useQuery } from "@tanstack/react-query";

import { useSessionApi } from "@application/shared/api";
import { type Session_GetProfile_Res } from "@application/shared/api/services";
import { QK } from "@application/shared/data/constants";

import { isSessionInvalidException } from "@infrastructure/api";

/**
 * Get profile query
 */
type GetProfileRes = Session_GetProfile_Res;

type GetProfileOptions = Omit<UseQueryOptions<GetProfileRes>, "queryKey" | "queryFn"> & {
    onSuccess?: (response: GetProfileRes) => void;
    onError?: (error: Error) => void;
    onSessionInvalid?: (error: Error) => void;
};

function useGetProfile({ onSuccess, onError, onSessionInvalid, ...options }: GetProfileOptions = {}) {
    // const { changeLanguage } = useI18n();

    const {
        queries: { getProfile },
    } = useSessionApi();

    async function queryFn({ signal }: QueryFunctionContext) {
        try {
            const res = await getProfile(signal);

            if (onSuccess) onSuccess(res);

            return res;
        } catch (error) {
            if (!(error instanceof Error)) {
                throw error;
            }

            if (isSessionInvalidException(error) && onSessionInvalid) {
                onSessionInvalid(error);

                throw error;
            }

            if (onError) onError(error);

            throw error;
        }
    }

    return useQuery({
        queryKey: [QK["session.get-profile"]],
        queryFn,

        ...options,
    });
}

export const SessionQueries = Object.freeze({
    useGetProfile,
});
