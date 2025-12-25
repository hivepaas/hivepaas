import type React from "react";

export interface GlobalAlertDialogProps {
    title?: React.ReactNode;
    description?: React.ReactNode;
    cancelText?: string;
    actionText?: string;
    onAction?: () => void;
    onCancel?: () => void;
    showFooter?: boolean;
    type?: "default" | "error";
}

export interface GlobalAlertDialogState {
    mode: "open" | "closed";
}

export interface GlobalAlertDialogOptions {
    props: GlobalAlertDialogProps;
}
