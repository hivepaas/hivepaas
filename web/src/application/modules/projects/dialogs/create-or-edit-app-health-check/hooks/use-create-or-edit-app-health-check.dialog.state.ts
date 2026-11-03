import { create } from "zustand";
import type { AppHealthCheck } from "~/projects/domain";

import type { CreateOrEditAppHealthCheckDialogOptions, CreateOrEditAppHealthCheckDialogState } from "../types";

type State = CreateOrEditAppHealthCheckDialogState & CreateOrEditAppHealthCheckDialogOptions;

interface Actions {
    open: (projectId: string, appId: string, options?: CreateOrEditAppHealthCheckDialogOptions) => void;
    openEdit: (
        projectId: string,
        appId: string,
        healthCheck: AppHealthCheck,
        options?: CreateOrEditAppHealthCheckDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditAppHealthCheckDialogState = create<State & Actions>()(set => ({
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
    openEdit: (projectId, appId, healthCheck, options = {}) => {
        set({
            state: {
                mode: "edit",
                projectId,
                appId,
                healthCheck,
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
