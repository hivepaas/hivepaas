import { create } from "zustand";
import type { AppConfigFile } from "~/projects/domain";

import type { CreateOrEditAppConfigFileDialogOptions, CreateOrEditAppConfigFileDialogState } from "../types";

type State = CreateOrEditAppConfigFileDialogState & CreateOrEditAppConfigFileDialogOptions;

interface Actions {
    open: (projectId: string, appId: string, options?: CreateOrEditAppConfigFileDialogOptions) => void;
    openEdit: (
        projectId: string,
        appId: string,
        configFile: AppConfigFile,
        options?: CreateOrEditAppConfigFileDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useCreateOrEditAppConfigFileDialogState = create<State & Actions>()(set => ({
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

    openEdit: (projectId, appId, configFile, options = {}) => {
        set({
            state: {
                mode: "edit",
                projectId,
                appId,
                configFile,
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
