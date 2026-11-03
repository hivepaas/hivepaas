import { create } from "zustand";

import { type CreateProjectAppDialogOptions, type CreateProjectAppDialogState } from "../types";

type State = CreateProjectAppDialogState & CreateProjectAppDialogOptions;

interface Actions {
    open: (projectId: string, options?: CreateProjectAppDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateProjectAppDialogState = create<State & Actions>()(set => ({
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
