import { create } from "zustand";

import type { RunNowTaskCreatedDialogOptions, RunNowTaskCreatedDialogState } from "../types";

type State = RunNowTaskCreatedDialogState & RunNowTaskCreatedDialogOptions;

interface Actions {
    open: (taskId: string, options?: RunNowTaskCreatedDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useRunNowTaskCreatedDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        taskId: null,
    },
    props: {},
    open: (taskId, options = {}) => {
        set({
            state: {
                mode: "open",
                taskId,
            },
            ...options,
        });
    },
    close: () => {
        set({
            state: {
                mode: "closed",
                taskId: null,
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
                    taskId: null,
                },
                props: {},
            };
        });
    },
}));
