import type React from "react";
import { useCallback } from "react";

import { type ExternalToast, toast } from "sonner";

import { useGlobalAlertDialog } from "@application/shared/dialogs";

import {
    isCancelException,
    isHighLevelException,
    isUnauthorizedException,
    isValidationException,
} from "@infrastructure/api/utils";

interface Params {
    message: React.ReactNode;
    error: Error;
    status?: "error" | "warning";
    options?: ExternalToast;
}

function createHook() {
    return function useApiErrorNotifications() {
        const { actions: globalAlertActions } = useGlobalAlertDialog();

        const notifyError = useCallback(
            ({ message, error, status = "error", options = {} }: Params) => {
                if (isCancelException(error)) {
                    return;
                }

                switch (true) {
                    case import.meta.env.MODE === "development":
                        console.error(error);

                        break;

                    default:
                        console.error(error.message);

                        break;
                }

                const toastOptions: ExternalToast = { ...options };

                if (isHighLevelException(error)) {
                    let errorMessage = error.message;
                    if (isValidationException(error)) {
                        const firstError = error.errors[0];
                        if (firstError) {
                            errorMessage = `Param '${firstError.path}': ${firstError.message}`;
                        }
                    }
                    globalAlertActions.open({
                        props: {
                            title: message,
                            description: errorMessage,
                            showFooter: false,
                            type: "error",
                        },
                    });

                    return;
                }

                if (isUnauthorizedException(error)) {
                    toast.error(message, {
                        description: "You don't have the appropriate permissions to perform this operation.",
                        ...toastOptions,
                    });

                    return;
                }

                if (isValidationException(error)) {
                    const firstError = error.errors[0];
                    if (firstError) {
                        const validationMessage = `Param '${firstError.path}': ${firstError.message}`;
                        toast.error(message, {
                            description: validationMessage,
                            ...toastOptions,
                        });

                        return;
                    }
                }

                if (status === "warning") {
                    toast.warning(message, {
                        description: error.message,
                        ...toastOptions,
                    });

                    return;
                }

                toast.error(message, {
                    description: error.message,
                    ...toastOptions,
                });
            },
            [globalAlertActions],
        );

        return {
            notifyError,
        };
    };
}

export const useApiErrorNotifications = createHook();
