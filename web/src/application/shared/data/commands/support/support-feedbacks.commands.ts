import { type UseMutationOptions, useMutation } from "@tanstack/react-query";

import { useSupportFeedbacksApi } from "@application/shared/api";
import type { SupportFeedbacks_CreateOne_Req, SupportFeedbacks_CreateOne_Res } from "@application/shared/api/services";

type CreateOneReq = SupportFeedbacks_CreateOne_Req["data"];
type CreateOneRes = SupportFeedbacks_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne(options: CreateOneOptions = {}) {
    const { mutations } = useSupportFeedbacksApi();

    return useMutation({
        mutationFn: mutations.createOne,
        ...options,
    });
}

export const SupportFeedbacksCommands = Object.freeze({
    useCreateOne,
});
