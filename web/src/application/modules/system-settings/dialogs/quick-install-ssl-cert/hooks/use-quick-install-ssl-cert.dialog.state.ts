import { create } from "zustand";

import type { QuickInstallSslCertDialogOptions, QuickInstallSslCertDialogState } from "../types";

type State = QuickInstallSslCertDialogState & QuickInstallSslCertDialogOptions;

interface Actions {
    open: (domain: string, options?: QuickInstallSslCertDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useQuickInstallSslCertDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
        domain: null,
    },
    props: {},

    open: (domain, options = {}) => {
        set({
            state: {
                mode: "open",
                domain,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
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
                    domain: null,
                },
                props: {},
            };
        });
    },
}));
