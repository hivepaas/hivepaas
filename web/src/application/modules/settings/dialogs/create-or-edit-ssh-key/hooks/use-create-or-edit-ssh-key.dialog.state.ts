import { create } from "zustand";
import type { SSHKeyTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditSSHKeyDialogOptions, CreateOrEditSSHKeyDialogState } from "../types";

type State = CreateOrEditSSHKeyDialogState & CreateOrEditSSHKeyDialogOptions;

interface Actions {
    open: (scope: SSHKeyTableScope, options?: CreateOrEditSSHKeyDialogOptions) => void;
    openEdit: (scope: SSHKeyTableScope, id: string, options?: CreateOrEditSSHKeyDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditSSHKeyDialogState = create<State & Actions>()(set => ({
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
