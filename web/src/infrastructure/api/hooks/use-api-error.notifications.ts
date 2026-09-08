import type React from "react";
import { useCallback } from "react";

import { type ExternalToast, toast } from "sonner";

import { useGlobalAlertDialog, useSettingInUseDialogState } from "@application/shared/dialogs";

import {
    isCancelException,
    isHighLevelException,
    isUnauthorizedException,
    isValidationException,
} from "@infrastructure/api/utils";

import { HttpException } from "@infrastructure/exceptions/http";

/**
 * A refusal that deserves a page of its own, not a toast.
 *
 * Handled here because every API hook funnels its failures through this one
 * function, and around fifty delete commands can raise it. Reacting to it at the
 * call sites would mean touching all of them and getting it wrong in one.
 */
const SETTING_IN_USE = "ERR_SETTING_IN_USE";

interface Params {
    message: React.ReactNode;
    error: Error;
    status?: "error" | "warning";
    options?: ExternalToast;
}

function createHook() {
    return function useApiErrorNotifications() {
        const { actions: globalAlertActions } = useGlobalAlertDialog();
        const settingInUseDialog = useSettingInUseDialogState();

        const notifyError = useCallback(
            ({ message, error, status = "error", options = {} }: Params) => {
                if (isCancelException(error)) {
                    return;
                }

                if (error instanceof HttpException && error.code === SETTING_IN_USE) {
                    // The failed request's URL is what locates the usage list -
                    // see HttpException.requestUrl.
                    settingInUseDialog.open({ props: { requestUrl: error.requestUrl } });
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
            [globalAlertActions, settingInUseDialog],
        );

        return {
            notifyError,
        };
    };
}

export const useApiErrorNotifications = createHook();
