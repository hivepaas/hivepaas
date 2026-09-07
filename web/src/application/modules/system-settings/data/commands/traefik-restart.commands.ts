import { type UseMutationOptions, useMutation } from "@tanstack/react-query";
import { useTraefikRestartApi } from "~/system-settings/api/hooks";
import type { TraefikRestart_Execute_Req, TraefikRestart_Execute_Res } from "~/system-settings/api/services";

type RestartReq = TraefikRestart_Execute_Req["data"] | undefined;
type RestartRes = TraefikRestart_Execute_Res;
type RestartOptions = Omit<UseMutationOptions<RestartRes, Error, RestartReq>, "mutationFn">;

function useRestart({ onSuccess, ...options }: RestartOptions = {}) {
    const { mutations } = useTraefikRestartApi();

    return useMutation({
        mutationFn: (data?: RestartReq) => mutations.execute(data ?? {}),
        onSuccess: (response, ...rest) => {
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const TraefikRestartCommands = Object.freeze({
    useRestart,
});
