import { create } from "zustand";
import type { RepoWebhookTableScope } from "~/settings/module-shared/components";

import type { UpdateRepoWebhookStatusDialogOptions, UpdateRepoWebhookStatusDialogState } from "../types";

type State = UpdateRepoWebhookStatusDialogState & UpdateRepoWebhookStatusDialogOptions;

interface Actions {
    open: (scope: RepoWebhookTableScope, id: string, options?: UpdateRepoWebhookStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateRepoWebhookStatusDialogState = create<State & Actions>()(set => ({
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
