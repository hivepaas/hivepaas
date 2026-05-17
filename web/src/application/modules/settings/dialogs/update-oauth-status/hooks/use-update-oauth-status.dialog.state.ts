import { create } from "zustand";

import type { UpdateOAuthStatusDialogOptions, UpdateOAuthStatusDialogState } from "../types";

type State = UpdateOAuthStatusDialogState & UpdateOAuthStatusDialogOptions;

interface Actions {
    open: (id: string, options?: UpdateOAuthStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateOAuthStatusDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: {},
    open: (id, options = {}) => {
        set({ state: { mode: "open", id }, ...options });
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
