import type React from "react";
import { useCallback } from "react";

import { type ExternalToast, toast } from "sonner";

import { isCancelException, isUnauthorizedException, isValidationException } from "@infrastructure/api/utils";

interface Params {
    message: React.ReactNode;
    error: Error;
    status?: "error" | "warning";
    options?: ExternalToast;
}

function createHook() {
    return function useApiErrorNotifications() {
        const notifyError = useCallback(({ message, error, status = "error", options = {} }: Params) => {
            if (isCancelException(error)) {
                return;
            }

            console.log(isValidationException(error));

            switch (true) {
                case import.meta.env.MODE === "development":
                    console.error(error);

                    break;

                default:
                    console.error(error.message);

                    break;
            }

            if (isUnauthorizedException(error)) {
                toast.error(message, {
                    description: "You don't have the appropriate permissions to perform this operation.",
                    ...options,
                });

                return;
            }

            if (isValidationException(error)) {
                console.log(error.errors);
                const firstError = error.errors[0];
                if (firstError) {
                    const validationMessage = `Param '${firstError.path}': ${firstError.message}`;
                    toast.error(message, {
                        description: validationMessage,
                        ...options,
                    });

                    return;
                }
            }

            if (status === "warning") {
                toast.warning(message, {
                    description: error.message,
                    ...options,
                });

                return;
            }

            toast.error(message, {
                description: error.message,
                ...options,
            });
        }, []);

        return {
            notifyError,
        };
    };
}

export const useApiErrorNotifications = createHook();
