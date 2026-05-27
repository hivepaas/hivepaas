import { create } from "zustand";
import type { GithubAppTableScope } from "~/settings/module-shared/components";

import type { ProvisionGithubAppDialogOptions, ProvisionGithubAppDialogState } from "../types";

type State = ProvisionGithubAppDialogState & ProvisionGithubAppDialogOptions;

interface Actions {
    open: (scope: GithubAppTableScope, options?: ProvisionGithubAppDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useProvisionGithubAppDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: undefined,
    open: (scope, options) => {
        set({ state: { mode: "open", scope }, props: options?.props });
    },
    close: () => {
        set(({ props }) => ({ state: { mode: "closed" }, props }));
    },
    clear: () => {
        set({ props: undefined });
    },
    destroy: () => {
        set({ state: { mode: "closed" }, props: undefined });
    },
}));
