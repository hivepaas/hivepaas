import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useMcpSettingsApi } from "~/system-settings/api/hooks";
import type { McpSettings_UpdateOne_Req, McpSettings_UpdateOne_Res } from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type UpdateOneReq = McpSettings_UpdateOne_Req["data"];
type UpdateOneRes = McpSettings_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useMcpSettingsApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({ queryKey: [QK["system-settings.mcp.find-one"]] });
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const McpSettingsCommands = Object.freeze({
    useUpdateOne,
});
