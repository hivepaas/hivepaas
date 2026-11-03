import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useEmailApi } from "~/settings/api/hooks";
import type {
    Email_CreateOne_Req,
    Email_CreateOne_Res,
    Email_DeleteOne_Req,
    Email_DeleteOne_Res,
    Email_TestSendMail_Req,
    Email_TestSendMail_Res,
    Email_UpdateOne_Req,
    Email_UpdateOne_Res,
    Email_UpdateStatus_Req,
    Email_UpdateStatus_Res,
} from "~/settings/api/services/email-services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = Email_CreateOne_Req["data"];
type CreateOneRes = Email_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useEmailApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.email.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = Email_UpdateOne_Req["data"];
type UpdateOneRes = Email_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useEmailApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.email.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.email.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = Email_UpdateStatus_Req["data"];
type UpdateStatusRes = Email_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useEmailApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.email.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.email.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = Email_DeleteOne_Req["data"];
type DeleteOneRes = Email_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useEmailApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.email.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.email.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type TestSendMailReq = Email_TestSendMail_Req["data"];
type TestSendMailRes = Email_TestSendMail_Res;
type TestSendMailOptions = Omit<UseMutationOptions<TestSendMailRes, Error, TestSendMailReq>, "mutationFn">;

function useTestSendMail(options: TestSendMailOptions = {}) {
    const { mutations } = useEmailApi();

    return useMutation({
        mutationFn: mutations.testSendMail,
        ...options,
    });
}

export const EmailCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
    useTestSendMail,
});
