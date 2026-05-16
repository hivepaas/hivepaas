import { create } from "zustand";
import type { AppStorageMount } from "~/projects/domain";

import type { StorageMountDialogOptions, StorageMountDialogState } from "../types";

type State = StorageMountDialogState & StorageMountDialogOptions;

interface Actions {
    open: (options?: StorageMountDialogOptions) => void;
    openEdit: (mount: AppStorageMount & { _id: string }, options?: StorageMountDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useStorageMountDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (options = {}) => {
        const { projectKey, appLocalKey, props = {} } = options;
        set({
            state: {
                mode: "open",
                projectKey,
                appLocalKey,
            },
            props,
        });
    },

    openEdit: (mount, options = {}) => {
        const { projectKey, appLocalKey, props = {} } = options;
        set({
            state: {
                mode: "edit",
                mount,
                projectKey,
                appLocalKey,
            },
            props,
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
