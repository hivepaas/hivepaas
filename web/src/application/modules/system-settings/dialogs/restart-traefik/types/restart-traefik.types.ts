export type RestartTraefikDialogMode = "open" | "closed";

export type RestartTraefikDialogState = {
    state: {
        mode: RestartTraefikDialogMode;
    };
};

export type RestartTraefikDialogOptions = {
    props?: {
        onClose?: () => void;
    };
};
