export interface InviteUserDialogState {
    state: { mode: "open" } | { mode: "closed" };
}

export interface InviteUserDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
