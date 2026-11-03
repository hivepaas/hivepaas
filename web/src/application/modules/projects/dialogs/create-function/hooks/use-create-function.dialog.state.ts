import { create } from "zustand";

import { type CreateFunctionDialogOptions, type CreateFunctionDialogState } from "../types";

type State = CreateFunctionDialogState & CreateFunctionDialogOptions;

interface Actions {
    open: (projectId: string, options?: CreateFunctionDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateFunctionDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
    },

    props: {},

    open: (projectId, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
                projectId: null,
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
                    projectId: null,
                },
                props: {},
            };
        });
    },
}));
