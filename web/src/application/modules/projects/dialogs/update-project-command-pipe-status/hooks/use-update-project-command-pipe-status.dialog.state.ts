import { create } from "zustand";

import type { UpdateProjectCommandPipeStatusDialogOptions, UpdateProjectCommandPipeStatusDialogState } from "../types";

type State = UpdateProjectCommandPipeStatusDialogState & UpdateProjectCommandPipeStatusDialogOptions;

interface Actions {
    open: (projectId: string, id: string, options?: UpdateProjectCommandPipeStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateProjectCommandPipeStatusDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (projectId, id, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
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
