import { type UseMutationOptions, useMutation } from "@tanstack/react-query";
import { useTraefikRestartApi } from "~/system-settings/api/hooks";
import type { TraefikRestart_Execute_Res } from "~/system-settings/api/services";

type RestartRes = TraefikRestart_Execute_Res;
type RestartOptions = Omit<UseMutationOptions<RestartRes>, "mutationFn">;

function useRestart({ onSuccess, ...options }: RestartOptions = {}) {
    const { mutations } = useTraefikRestartApi();

    return useMutation({
        mutationFn: () => mutations.execute({}),
        onSuccess: (response, ...rest) => {
            onSuccess?.(response, ...rest);
        },
        ...options,
    });
}

export const TraefikRestartCommands = Object.freeze({
    useRestart,
});
