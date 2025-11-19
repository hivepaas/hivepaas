import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useUsersApi } from "~/user-management/api/hooks";
import {
    type Users_DeleteOne_Req,
    type Users_DeleteOne_Res,
    type Users_UpdateOne_Req,
    type Users_UpdateOne_Res,
} from "~/user-management/api/services";
import { QK } from "~/user-management/data/constants";

/**
 * Delete a user command
 */
type DeleteOneReq = Users_DeleteOne_Req["data"];
type DeleteOneRes = Users_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useUsersApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            // const { id } = response.data;

            queryClient.removeQueries({
                queryKey: [QK["users.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },

        ...options,
    });
}

/**
 * Update a user command
 */
type UpdateOneReq = Users_UpdateOne_Req["data"];
type UpdateOneRes = Users_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useUsersApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["users.$.find-many-paginated"]],
            });

            void queryClient.invalidateQueries({
                queryKey: [QK["users.$.find-one-by-id"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },

        ...options,
    });
}

export const UsersCommands = Object.freeze({
    useDeleteOne,
    useUpdateOne,
});
