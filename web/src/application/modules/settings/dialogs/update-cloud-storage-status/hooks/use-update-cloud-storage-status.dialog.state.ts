import { create } from "zustand";
import type { CloudStorageTableScope } from "~/settings/module-shared/components";

import type { UpdateCloudStorageStatusDialogOptions, UpdateCloudStorageStatusDialogState } from "../types";

type State = UpdateCloudStorageStatusDialogState & UpdateCloudStorageStatusDialogOptions;

interface Actions {
    open: (scope: CloudStorageTableScope, id: string, options?: UpdateCloudStorageStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateCloudStorageStatusDialogState = create<State & Actions>()(set => ({
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
