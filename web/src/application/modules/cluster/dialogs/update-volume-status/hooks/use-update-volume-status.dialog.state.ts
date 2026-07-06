import { create } from "zustand";
import type { VolumeManagementScope } from "~/cluster/module-shared/types";

import type { UpdateVolumeStatusDialogOptions, UpdateVolumeStatusDialogState } from "../types";

type State = UpdateVolumeStatusDialogState & UpdateVolumeStatusDialogOptions;

interface Actions {
    open: (scope: VolumeManagementScope, id: string, options?: UpdateVolumeStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateVolumeStatusDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (scope, id, options = {}) => {
        set({
            state: {
                mode: "open",
                scope,
                id,
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
