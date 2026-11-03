export type ChangeKekDialogMode = "open" | "closed";

export type ChangeKekDialogState = {
    state: {
        mode: ChangeKekDialogMode;
    };
};

export type ChangeKekDialogOptions = {
    props?: {
        onClose?: () => void;
    };
};
