export interface CreateFeedbackDialogState {
    state: {
        mode: "closed" | "open";
    };
}

export interface CreateFeedbackDialogOptions {
    props?: {
        onClose?: () => void;
    };
}
