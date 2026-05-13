import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useImServiceApi } from "~/settings/api/hooks";
import type {
    ImService_CreateOne_Req,
    ImService_CreateOne_Res,
    ImService_DeleteOne_Req,
    ImService_DeleteOne_Res,
    ImService_TestSendMsg_Req,
    ImService_TestSendMsg_Res,
    ImService_UpdateOne_Req,
    ImService_UpdateOne_Res,
    ImService_UpdateStatus_Req,
    ImService_UpdateStatus_Res,
} from "~/settings/api/services/im-service-services";
import { QK } from "~/settings/data/constants";

type CreateOneReq = ImService_CreateOne_Req["data"];
type CreateOneRes = ImService_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useImServiceApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.im-service.find-many-paginated"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = ImService_UpdateOne_Req["data"];
type UpdateOneRes = ImService_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useImServiceApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.im-service.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.im-service.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = ImService_UpdateStatus_Req["data"];
type UpdateStatusRes = ImService_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useImServiceApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.im-service.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.im-service.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ImService_DeleteOne_Req["data"];
type DeleteOneRes = ImService_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useImServiceApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.im-service.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["settings.im-service.find-one-by-id"]],
            });

            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

type TestSendMsgReq = ImService_TestSendMsg_Req["data"];
type TestSendMsgRes = ImService_TestSendMsg_Res;
type TestSendMsgOptions = Omit<UseMutationOptions<TestSendMsgRes, Error, TestSendMsgReq>, "mutationFn">;

function useTestSendMsg(options: TestSendMsgOptions = {}) {
    const { mutations } = useImServiceApi();

    return useMutation({
        mutationFn: mutations.testSendMsg,
        ...options,
    });
}

export const ImServiceCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
    useTestSendMsg,
});
