import { create } from "zustand";

import type { ExportContainerFilesDialogOptions, ExportContainerFilesDialogState } from "../types";

type State = ExportContainerFilesDialogState & ExportContainerFilesDialogOptions;

interface Actions {
    open: (
        projectId: string,
        env: string,
        appId: string,
        nodeId: string,
        containerId: string,
        options?: ExportContainerFilesDialogOptions,
    ) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useExportContainerFilesDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        env: null,
        appId: null,
        nodeId: null,
        containerId: null,
    },
    props: {},
    open: (projectId, env, appId, nodeId, containerId, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
                env,
                appId,
                nodeId,
                containerId,
            },
            ...options,
        });
    },
    close: () => {
        set({
            state: {
                mode: "closed",
                projectId: null,
                env: null,
                appId: null,
                nodeId: null,
                containerId: null,
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
                    env: null,
                    appId: null,
                    nodeId: null,
                    containerId: null,
                },
                props: {},
            };
        });
    },
}));
