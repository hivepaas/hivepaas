import { create } from "zustand";

import type { DockerApiPermissionsGuideDialogOptions, DockerApiPermissionsGuideDialogState } from "../types";

type State = DockerApiPermissionsGuideDialogState & DockerApiPermissionsGuideDialogOptions;

interface Actions {
    open: (options?: DockerApiPermissionsGuideDialogOptions) => void;
    close: () => void;
    clear: () => void;
    destroy: () => void;
}

export const useDockerApiPermissionsGuideDialogState = create<State & Actions>()(set => ({
    state: { mode: "closed" },
    props: {},
    open: (options = {}) => {
        set({ state: { mode: "open" }, props: options.props ?? {} });
    },
    close: () => {
        set({ state: { mode: "closed" }, props: {} });
    },
    clear: () => {
        set({ props: {} });
    },
    destroy: () => {
        set(state => {
            if (state.state.mode === "closed") {
                return state;
            }
            return { state: { mode: "closed" }, props: {} };
        });
    },
}));
