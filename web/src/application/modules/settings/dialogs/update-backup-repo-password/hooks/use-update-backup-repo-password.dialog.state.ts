import { create } from "zustand";
import type { BackupRepoTableScope } from "~/settings/module-shared/components";

import type { UpdateBackupRepoPasswordDialogOptions, UpdateBackupRepoPasswordDialogState } from "../types";

type State = UpdateBackupRepoPasswordDialogState & UpdateBackupRepoPasswordDialogOptions;

interface Actions {
    open: (
        scope: BackupRepoTableScope,
        id: string,
        updateVer?: number,
        options?: UpdateBackupRepoPasswordDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateBackupRepoPasswordDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: {},
    open: (scope, id, updateVer, options = {}) => {
        set({ state: { mode: "open", scope, id, updateVer }, ...options });
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
