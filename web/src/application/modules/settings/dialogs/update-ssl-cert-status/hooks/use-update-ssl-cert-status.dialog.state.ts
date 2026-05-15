import { create } from "zustand";
import type { SslCertTableScope } from "~/settings/module-shared/components";

import type { UpdateSslCertStatusDialogOptions, UpdateSslCertStatusDialogState } from "../types";

type State = UpdateSslCertStatusDialogState & UpdateSslCertStatusDialogOptions;

interface Actions {
    open: (scope: SslCertTableScope, id: string, options?: UpdateSslCertStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateSslCertStatusDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },

    props: {},

    open: (scope, id, options = {}) => {
        set({
            state: {
                mode: "open",
                scope,
                id,
            },
            ...options,
        });
    },

    close: () => {
        set({
            state: {
                mode: "closed",
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
                },
                props: {},
            };
        });
    },
}));
