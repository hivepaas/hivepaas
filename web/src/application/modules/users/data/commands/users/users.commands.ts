import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useUsersApi } from "~/users/api/hooks";
import { type Users_DeleteOne_Req, type Users_DeleteOne_Res } from "~/users/api/services";
import { QK } from "~/users/data/constants";

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

export const UsersCommands = Object.freeze({
    useDeleteOne,
});
