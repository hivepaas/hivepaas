import { create } from "zustand";

import { type CopyProjectAppDialogOptions, type CopyProjectAppDialogState } from "../types";

type State = CopyProjectAppDialogState & CopyProjectAppDialogOptions;

interface Actions {
    open: (projectId: string, appId: string, appEnv: string, options?: CopyProjectAppDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCopyProjectAppDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        appId: null,
    },

    props: {},

    open: (projectId, appId, appEnv, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
                appId,
                appEnv,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
                projectId: null,
                appId: null,
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
                    appId: null,
                },
                props: {},
            };
        });
    },
}));
