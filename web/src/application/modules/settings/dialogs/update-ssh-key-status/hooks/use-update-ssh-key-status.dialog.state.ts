import { create } from "zustand";
import type { SSHKeyTableScope } from "~/settings/module-shared/components";

import type { UpdateSSHKeyStatusDialogOptions, UpdateSSHKeyStatusDialogState } from "../types";

type State = UpdateSSHKeyStatusDialogState & UpdateSSHKeyStatusDialogOptions;

interface Actions {
    open: (scope: SSHKeyTableScope, id: string, options?: UpdateSSHKeyStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateSSHKeyStatusDialogState = create<State & Actions>()(set => ({
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
