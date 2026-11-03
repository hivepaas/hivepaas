import { type UseMutationOptions, useMutation } from "@tanstack/react-query";
import { useHivePaaSRestartApi } from "~/system-settings/api/hooks";
import type { HivePaaSRestart_Execute_Req, HivePaaSRestart_Execute_Res } from "~/system-settings/api/services";

type RestartReq = HivePaaSRestart_Execute_Req["data"];
type RestartRes = HivePaaSRestart_Execute_Res;
type RestartOptions = Omit<UseMutationOptions<RestartRes, Error, RestartReq>, "mutationFn">;

function useRestart({ onSuccess, ...options }: RestartOptions = {}) {
    const { mutations } = useHivePaaSRestartApi();

    return useMutation({
        mutationFn: mutations.execute,
        onSuccess: (response, ...rest) => {
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const HivePaaSRestartCommands = Object.freeze({
    useRestart,
});
