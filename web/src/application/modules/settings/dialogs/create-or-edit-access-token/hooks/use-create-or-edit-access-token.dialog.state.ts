import { create } from "zustand";
import type { AccessTokenTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditAccessTokenDialogOptions, CreateOrEditAccessTokenDialogState } from "../types";

type State = CreateOrEditAccessTokenDialogState & CreateOrEditAccessTokenDialogOptions;

interface Actions {
    open: (scope: AccessTokenTableScope, options?: CreateOrEditAccessTokenDialogOptions) => void;
    openEdit: (scope: AccessTokenTableScope, id: string, options?: CreateOrEditAccessTokenDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditAccessTokenDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: {},
    open: (scope, options = {}) => {
        set({ state: { mode: "open", scope }, ...options });
    },
    openEdit: (scope, id, options = {}) => {
        set({ state: { mode: "edit", scope, id }, ...options });
    },
    close: () => {
        set({ state: { mode: "closed" } });
    },
    clear: () => {
        set({ props: {} });
    },
    destroy: () => {
        set(state => {
            if (state.state.mode === "closed") return state;
            return { state: { mode: "closed" }, props: {} };
        });
    },
}));
