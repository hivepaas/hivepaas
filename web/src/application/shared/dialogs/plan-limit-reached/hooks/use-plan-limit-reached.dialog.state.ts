import { create } from "zustand";

import type { PlanLimitReachedDialogOptions, PlanLimitReachedDialogState } from "../types";

type State = PlanLimitReachedDialogState & PlanLimitReachedDialogOptions;

interface Actions {
    open: (options?: PlanLimitReachedDialogOptions) => void;
    close: () => void;
    destroy: () => void;
}

export const usePlanLimitReachedDialogState = create<State & Actions>()(set => ({
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
