export interface UpdateOAuthStatusDialogState {
    state: { mode: "open"; id: string } | { mode: "closed" };
}

export interface UpdateOAuthStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
