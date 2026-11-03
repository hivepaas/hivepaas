import { create } from "zustand";
import type { GithubAppTableScope } from "~/settings/module-shared/components";

import type { UpdateGithubAppStatusDialogOptions, UpdateGithubAppStatusDialogState } from "../types";

type State = UpdateGithubAppStatusDialogState & UpdateGithubAppStatusDialogOptions;

interface Actions {
    open: (scope: GithubAppTableScope, id: string, options?: UpdateGithubAppStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateGithubAppStatusDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: undefined,
    open: (scope, id, options) => {
        set({ state: { mode: "open", scope, id }, props: options?.props });
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
