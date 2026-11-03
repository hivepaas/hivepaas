export type RestartHivePaaSDialogMode = "open" | "closed";

export type RestartHivePaaSDialogState = {
    state: {
        mode: RestartHivePaaSDialogMode;
    };
};

export type RestartHivePaaSDialogOptions = {
    props?: {
        onClose?: () => void;
    };
};
