import { create } from "zustand";
import type { GithubAppTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditGithubAppDialogOptions, CreateOrEditGithubAppDialogState } from "../types";

type State = CreateOrEditGithubAppDialogState & CreateOrEditGithubAppDialogOptions;

interface Actions {
    open: (scope: GithubAppTableScope, options?: CreateOrEditGithubAppDialogOptions) => void;
    openEdit: (scope: GithubAppTableScope, id: string, options?: CreateOrEditGithubAppDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditGithubAppDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: undefined,
    open: (scope, options) => {
        set({ state: { mode: "open", scope }, props: options?.props });
    },
    openEdit: (scope, id, options) => {
        set({ state: { mode: "edit", scope, id }, props: options?.props });
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
