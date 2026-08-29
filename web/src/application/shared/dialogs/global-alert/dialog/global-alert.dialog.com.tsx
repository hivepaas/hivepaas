import React from "react";

import { Button } from "@components/ui";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { X } from "lucide-react";

import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from "@/components/ui/alert-dialog";

import { useGlobalAlertDialogState } from "../hooks/use-global-alert.dialog.state";

const fnPlaceholder = () => null;

function View() {
    const {
        mode,
        props: {
            title = "",
            description = "",
            actionText = "OK",
            cancelText = "Cancel",
            onAction = fnPlaceholder,
            onCancel = fnPlaceholder,
            showFooter = true,
            type = "default",
        } = {},
        close,
    } = useGlobalAlertDialogState();

    const onOpenChange = (open: boolean) => {
        if (!open) {
            close();
        }
    };

    return (
        <AlertDialog
            open={mode === "open"}
            onOpenChange={onOpenChange}
        >
            <AlertDialogContent className="w-full max-w-[calc(100vw-1.5rem)] sm:max-w-[500px] p-2.5 sm:p-4.5 gap-2 sm:gap-3.5 overflow-hidden">
                <AlertDialogHeader className="min-w-0 gap-1.5 sm:gap-2.5">
                    <AlertDialogTitle className="pr-7 sm:pr-8 text-base sm:text-lg text-center">
                        {title}
                    </AlertDialogTitle>
                    <AlertDialogDescription
                        className={cn(
                            dashedBorderBox,
                            type === "error"
                                ? "text-red-500"
                                : type === "warning"
                                  ? "text-yellow-500"
                                  : "text-gray-500",
                            "px-2.5 py-1.5 sm:px-3.5 sm:py-2 text-xs sm:text-sm min-w-0 max-w-full max-h-[min(50vh,24rem)] overflow-y-auto whitespace-pre-wrap break-words break-all",
                        )}
                    >
                        {description}
                    </AlertDialogDescription>
                    <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={close}
                        className="absolute right-2 top-2 sm:right-3.5 sm:top-3.5 size-7 sm:size-8"
                    >
                        <X className="size-4" />
                    </Button>
                </AlertDialogHeader>

                {showFooter ? (
                    <AlertDialogFooter className="gap-2 sm:gap-3">
                        {cancelText && <AlertDialogCancel onClick={onCancel}>{cancelText}</AlertDialogCancel>}
                        <AlertDialogAction onClick={onAction}>{actionText}</AlertDialogAction>
                    </AlertDialogFooter>
                ) : null}
            </AlertDialogContent>
        </AlertDialog>
    );
}

export const GlobalAlertDialog = React.memo(View);
