import { type UseMutationOptions, useMutation } from "@tanstack/react-query";
import { useHivePaaSAppSecretApi } from "~/system-settings/api/hooks";
import type { HivePaaSAppSecret_UpdateOne_Req, HivePaaSAppSecret_UpdateOne_Res } from "~/system-settings/api/services";

type UpdateOneReq = HivePaaSAppSecret_UpdateOne_Req["data"];
type UpdateOneRes = HivePaaSAppSecret_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useHivePaaSAppSecretApi();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const HivePaaSAppSecretCommands = Object.freeze({
    useUpdateOne,
});
