export interface DockerApiPermissionsGuideDialogState {
    state: { mode: "open" } | { mode: "closed" };
}

export interface DockerApiPermissionsGuideDialogOptions {
    props?: {
        onClose?: () => void;
    };
}
