import { create } from "zustand";

import type { CreateOrEditOAuthDialogOptions, CreateOrEditOAuthDialogState } from "../types";

type State = CreateOrEditOAuthDialogState & CreateOrEditOAuthDialogOptions;

interface Actions {
    open: (options?: CreateOrEditOAuthDialogOptions) => void;
    openEdit: (id: string, options?: CreateOrEditOAuthDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditOAuthDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: {},
    open: (options = {}) => {
        set({ state: { mode: "open" }, ...options });
    },
    openEdit: (id, options = {}) => {
        set({ state: { mode: "edit", id }, ...options });
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
