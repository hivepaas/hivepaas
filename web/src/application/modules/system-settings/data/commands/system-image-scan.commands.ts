import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useSystemImageScanApi } from "~/system-settings/api/hooks";
import type {
    SystemImageScan_Execute_Res,
    SystemImageScan_UpdateOne_Req,
    SystemImageScan_UpdateOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = SystemImageScan_UpdateOne_Req["data"];
type UpdateOneRes = SystemImageScan_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;
type ExecuteOptions = Omit<UseMutationOptions<SystemImageScan_Execute_Res>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useSystemImageScanApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.image-scan.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

function useExecute({ onSuccess, ...options }: ExecuteOptions = {}) {
    const { mutations } = useSystemImageScanApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.execute,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.image-scan.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const SystemImageScanCommands = Object.freeze({
    useUpdateOne,
    useExecute,
});
