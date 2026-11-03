import { create } from "zustand";
import type { SslProviderTableScope } from "~/settings/module-shared/components";

import type { UpdateSslProviderStatusDialogOptions, UpdateSslProviderStatusDialogState } from "../types";

type State = UpdateSslProviderStatusDialogState & UpdateSslProviderStatusDialogOptions;

interface Actions {
    open: (scope: SslProviderTableScope, id: string, options?: UpdateSslProviderStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateSslProviderStatusDialogState = create<State & Actions>()(set => ({
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
