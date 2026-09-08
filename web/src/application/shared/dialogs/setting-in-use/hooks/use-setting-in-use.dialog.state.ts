import { create } from "zustand";

import type { SettingInUseDialogOptions, SettingInUseDialogState } from "../types";

type State = SettingInUseDialogState & SettingInUseDialogOptions;

interface Actions {
    open: (options?: SettingInUseDialogOptions) => void;
    close: () => void;
    destroy: () => void;
}

export const useSettingInUseDialogState = create<State & Actions>()(set => ({
    mode: "closed",
    props: {},

    open: (options = { props: {} }) => {
        set({ mode: "open", ...options });
    },

    close: () => {
        set({ mode: "closed" });
    },

    destroy: () => {
        set(state => (state.mode === "closed" ? state : { mode: "closed", props: {} }));
    },
}));
