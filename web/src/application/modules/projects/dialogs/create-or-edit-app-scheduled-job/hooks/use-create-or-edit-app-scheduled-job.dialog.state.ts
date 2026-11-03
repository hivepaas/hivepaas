import { create } from "zustand";

import type { CreateOrEditAppScheduledJobDialogOptions, CreateOrEditAppScheduledJobDialogState } from "../types";

type State = CreateOrEditAppScheduledJobDialogState & CreateOrEditAppScheduledJobDialogOptions;

interface Actions {
    open: (projectId: string, appId: string, options?: CreateOrEditAppScheduledJobDialogOptions) => void;
    openEdit: (
        projectId: string,
        appId: string,
        scheduledJobId: string,
        options?: CreateOrEditAppScheduledJobDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditAppScheduledJobDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        appId: null,
    },
    props: {},
    open: (projectId, appId, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
                appId,
            },
            ...options,
        });
    },
    openEdit: (projectId, appId, scheduledJobId, options = {}) => {
        set({
            state: {
                mode: "edit",
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
