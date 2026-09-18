import { create } from "zustand";

import { type DeployTemplateDialogOptions, type DeployTemplateDialogState } from "../types";

type State = DeployTemplateDialogState;

interface Actions {
    open: (projectId: string, options: DeployTemplateDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useDeployTemplateDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        templateName: null,
    },

    props: {
        templateName: "",
    },

    open: (projectId, options) => {
        set({
            state: {
                mode: "open",
                projectId,
                templateName: options.templateName,
            },
            props: options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
                projectId: null,
                templateName: null,
            },
            props: {
                templateName: "",
            },
        });
    },

    clear: () => {
        set({
            props: {
                templateName: "",
            },
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
                    templateName: null,
                },
                props: {
                    templateName: "",
                },
            };
        });
    },
}));
