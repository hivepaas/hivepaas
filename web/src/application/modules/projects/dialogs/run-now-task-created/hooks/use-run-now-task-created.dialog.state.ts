import { create } from "zustand";

import type { RunNowTaskCreatedDialogOptions, RunNowTaskCreatedDialogState } from "../types";

type State = RunNowTaskCreatedDialogState & RunNowTaskCreatedDialogOptions;

interface Actions {
    open: (
        projectId: string,
        appId: string,
        scheduledJobId: string,
        taskId: string,
        options?: RunNowTaskCreatedDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useRunNowTaskCreatedDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        appId: null,
        scheduledJobId: null,
        taskId: null,
    },
    props: {},
    open: (projectId, appId, scheduledJobId, taskId, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
                appId,
                scheduledJobId,
                taskId,
            },
            props: options.props ?? {},
        });
    },
    close: () => {
        set({
            state: {
                mode: "closed",
                projectId: null,
                appId: null,
                scheduledJobId: null,
                taskId: null,
            },
            props: {},
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
                    taskId: null,
                },
                props: {},
            };
        });
    },
}));
