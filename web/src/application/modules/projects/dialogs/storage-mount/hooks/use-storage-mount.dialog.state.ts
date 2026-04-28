import { create } from "zustand";
import type { AppStorageMount, ProjectStorageSettings } from "~/projects/domain";

import type { StorageMountDialogOptions, StorageMountDialogState } from "../types";

type State = StorageMountDialogState & StorageMountDialogOptions;

interface Actions {
    open: (projectRules?: ProjectStorageSettings, options?: StorageMountDialogOptions) => void;
    openEdit: (
        mount: AppStorageMount & { _id: string },
        projectRules?: ProjectStorageSettings,
        options?: StorageMountDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useStorageMountDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (projectRules, options = {}) => {
        set({
            state: {
                mode: "open",
                projectRules,
            },
            ...options,
        });
    },

    openEdit: (mount, projectRules, options = {}) => {
        set({
            state: {
                mode: "edit",
                mount,
                projectRules,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
            },
        });
    },

    clear: () => {
        set({
            props: {},
        });
    },

    destroy: () => {
        set(state => {
            if (state.state.mode === "closed") {
                return state;
            }

            return {
                state: {
                    mode: "closed",
                },
                props: {},
            };
        });
    },
}));
