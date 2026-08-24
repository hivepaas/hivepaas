import { create } from "zustand";

import type { QuickInstallSslCertDialogOptions, QuickInstallSslCertDialogState } from "../types";

type State = QuickInstallSslCertDialogState & QuickInstallSslCertDialogOptions;

interface Actions {
    open: (projectId: string, env: string, domain: string, options?: QuickInstallSslCertDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useQuickInstallSslCertDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        projectId: null,
        env: null,
        domain: null,
    },
    props: {},

    open: (projectId, env, domain, options = {}) => {
        set({
            state: {
                mode: "open",
                projectId,
                env,
                domain,
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
                domain: null,
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
                    domain: null,
                },
                props: {},
            };
        });
    },
}));
