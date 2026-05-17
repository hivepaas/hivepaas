import type { SSHKeyTableScope } from "~/settings/module-shared/components";

export interface UpdateSSHKeyStatusDialogState {
    state: { mode: "open"; scope: SSHKeyTableScope; id: string } | { mode: "closed" };
}

export interface UpdateSSHKeyStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
