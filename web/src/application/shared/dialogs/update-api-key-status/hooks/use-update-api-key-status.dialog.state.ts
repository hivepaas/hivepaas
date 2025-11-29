import { create } from "zustand";

import type { ProfileApiKey } from "@application/shared/entities/profile";

import {
    type UpdateApiKeyStatusDialogOptions,
    type UpdateApiKeyStatusDialogState,
} from "../types/update-api-key-status.dialog.type";

type State = UpdateApiKeyStatusDialogState & UpdateApiKeyStatusDialogOptions;

interface Actions {
    open: (apiKey: ProfileApiKey, options?: UpdateApiKeyStatusDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useUpdateApiKeyStatusDialogState = create<State & Actions>()(set => ({
    state: {
        mode: "closed",
    },
    props: {},

    open: (apiKey, options = {}) => {
        set({
            state: {
                mode: "open",
                apiKey,
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

