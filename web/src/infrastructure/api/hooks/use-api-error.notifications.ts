import type React from "react";
import { useCallback } from "react";

import { type ExternalToast, toast } from "sonner";

import { isCancelException, isUnauthorizedException } from "@infrastructure/api/utils";

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

            switch (true) {
                case import.meta.env.MODE === "development":
                    console.error(error.stack);

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
