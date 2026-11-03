import { create } from "zustand";

import type { UpdateAppScheduledJobStatusDialogOptions, UpdateAppScheduledJobStatusDialogState } from "../types";

type State = UpdateAppScheduledJobStatusDialogState & UpdateAppScheduledJobStatusDialogOptions;

interface Actions {
    open: (
        projectId: string,
        appId: string,
        scheduledJobId: string,
        options?: UpdateAppScheduledJobStatusDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateAppScheduledJobStatusDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        appId: null,
        scheduledJobId: null,
    },
    props: {},
    open: (projectId, appId, scheduledJobId, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
                appId,
                scheduledJobId,
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
                scheduledJobId: null,
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
                    scheduledJobId: null,
                },
                props: {},
            };
        });
    },
}));
