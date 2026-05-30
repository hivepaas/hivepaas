import { create } from "zustand";

import type { UpdateAppHealthCheckStatusDialogOptions, UpdateAppHealthCheckStatusDialogState } from "../types";

type State = UpdateAppHealthCheckStatusDialogState & UpdateAppHealthCheckStatusDialogOptions;

interface Actions {
    open: (
        projectId: string,
        appId: string,
        healthCheckId: string,
        options?: UpdateAppHealthCheckStatusDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateAppHealthCheckStatusDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        appId: null,
        healthCheckId: null,
    },
    props: {},
    open: (projectId, appId, healthCheckId, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
                appId,
                healthCheckId,
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
                healthCheckId: null,
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
                    healthCheckId: null,
                },
                props: {},
            };
        });
    },
}));
