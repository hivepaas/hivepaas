import type { SSHKeyTableScope } from "~/settings/module-shared/components";

export interface CreateOrEditSSHKeyDialogState {
    state:
        | { mode: "open"; scope: SSHKeyTableScope }
        | { mode: "edit"; scope: SSHKeyTableScope; id: string }
        | { mode: "closed" };
}

export interface CreateOrEditSSHKeyDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
        readOnlyInherited?: boolean;
        entityTitle?: string;
    };
}
