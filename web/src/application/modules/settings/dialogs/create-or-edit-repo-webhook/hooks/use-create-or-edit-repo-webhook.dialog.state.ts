import { create } from "zustand";
import type { RepoWebhookTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditRepoWebhookDialogOptions, CreateOrEditRepoWebhookDialogState } from "../types";

type State = CreateOrEditRepoWebhookDialogState & CreateOrEditRepoWebhookDialogOptions;

interface Actions {
    open: (scope: RepoWebhookTableScope, options?: CreateOrEditRepoWebhookDialogOptions) => void;
    openEdit: (scope: RepoWebhookTableScope, id: string, options?: CreateOrEditRepoWebhookDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditRepoWebhookDialogState = create<State & Actions>()(set => ({
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
