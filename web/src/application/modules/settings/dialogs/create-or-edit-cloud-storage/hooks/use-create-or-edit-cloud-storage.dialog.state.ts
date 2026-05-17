import { create } from "zustand";
import type { CloudStorageTableScope } from "~/settings/module-shared/components";

import type { CreateOrEditCloudStorageDialogOptions, CreateOrEditCloudStorageDialogState } from "../types";

type State = CreateOrEditCloudStorageDialogState & CreateOrEditCloudStorageDialogOptions;

interface Actions {
    open: (scope: CloudStorageTableScope, options?: CreateOrEditCloudStorageDialogOptions) => void;
    openEdit: (scope: CloudStorageTableScope, id: string, options?: CreateOrEditCloudStorageDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditCloudStorageDialogState = create<State & Actions>()(set => ({
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
