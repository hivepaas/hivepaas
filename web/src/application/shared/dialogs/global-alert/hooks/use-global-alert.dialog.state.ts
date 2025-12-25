import { create } from "zustand";

import { type GlobalAlertDialogOptions, type GlobalAlertDialogState } from "../types/global-alert.dialog.type";

type State = GlobalAlertDialogState & GlobalAlertDialogOptions;

interface Actions {
    open: (options?: GlobalAlertDialogOptions) => void;
    close: () => void;
    destroy: () => void;
}

export const useGlobalAlertDialogState = create<State & Actions>()(set => ({
    mode: "closed",
    props: {},

    open: (options = { props: {} }) => {
        set({
            mode: "open",
            ...options,
        });
    },

    close: () => {
        set({
            mode: "closed",
        });
    },

    destroy: () => {
        set(state => {
            if (state.mode === "closed") {
                return state;
            }

            return {
                mode: "closed",
                props: {},
            };
        });
    },
}));
