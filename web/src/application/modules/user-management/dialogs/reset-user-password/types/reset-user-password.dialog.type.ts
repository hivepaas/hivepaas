import { type UserBase } from "~/user-management/domain";

export interface ResetUserPasswordDialogState {
    state:
        | {
              mode: "open";
              user: UserBase;
          }
        | {
              mode: "closed";
              user: null;
          };
}

export interface ResetUserPasswordDialogOptions {
    props?: {
        onSuccess?: (resetPasswordLink: string) => void;
        onError?: (error: Error) => void;
        onClose?: () => void;
    };
}
