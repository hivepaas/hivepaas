import type { ProfileApiKey } from "@application/shared/entities/profile";

export interface UpdateApiKeyStatusDialogState {
    state:
        | {
              mode: "open";
              apiKey: ProfileApiKey;
          }
        | { mode: "closed" };
}

export interface UpdateApiKeyStatusDialogOptions {
    props?: {
        onClose?: () => void;
        onSuccess?: () => void;
        onError?: (error: Error) => void;
    };
}
