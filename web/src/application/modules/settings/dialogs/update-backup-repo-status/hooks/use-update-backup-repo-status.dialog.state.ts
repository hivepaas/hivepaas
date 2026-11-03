import { create } from "zustand";
import type { BackupRepoTableScope } from "~/settings/module-shared/components";

import type { UpdateBackupRepoStatusDialogOptions, UpdateBackupRepoStatusDialogState } from "../types";

type State = UpdateBackupRepoStatusDialogState & UpdateBackupRepoStatusDialogOptions;

interface Actions {
    open: (scope: BackupRepoTableScope, id: string, options?: UpdateBackupRepoStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateBackupRepoStatusDialogState = create<State & Actions>()(set => ({
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
