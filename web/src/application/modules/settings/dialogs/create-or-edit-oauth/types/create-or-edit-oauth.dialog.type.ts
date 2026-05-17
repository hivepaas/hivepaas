export interface CreateOrEditOAuthDialogState {
    state: { mode: "open" } | { mode: "edit"; id: string } | { mode: "closed" };
}

export interface CreateOrEditOAuthDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
