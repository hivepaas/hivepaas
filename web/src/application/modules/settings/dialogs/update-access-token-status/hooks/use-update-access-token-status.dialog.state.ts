import { create } from "zustand";
import type { AccessTokenTableScope } from "~/settings/module-shared/components";

import type { UpdateAccessTokenStatusDialogOptions, UpdateAccessTokenStatusDialogState } from "../types";

type State = UpdateAccessTokenStatusDialogState & UpdateAccessTokenStatusDialogOptions;

interface Actions {
    open: (scope: AccessTokenTableScope, id: string, options?: UpdateAccessTokenStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateAccessTokenStatusDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: {},
    open: (scope, id, options = {}) => {
        set({ state: { mode: "open", scope, id }, ...options });
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
