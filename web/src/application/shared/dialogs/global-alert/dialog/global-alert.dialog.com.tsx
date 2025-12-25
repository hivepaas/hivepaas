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
            <AlertDialogContent>
                <AlertDialogHeader>
                    <AlertDialogTitle>{title}</AlertDialogTitle>
                    <AlertDialogDescription
                        className={cn("text-", type === "error" ? "text-red-500" : "text-gray-500")}
                    >
                        <div className={cn(dashedBorderBox)}>{description}</div>
                    </AlertDialogDescription>
                    <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={close}
                        className="absolute right-4 top-4"
                    >
                        <X />
                    </Button>
                </AlertDialogHeader>

                {showFooter ? (
                    <AlertDialogFooter>
                        {cancelText && <AlertDialogCancel onClick={onCancel}>{cancelText}</AlertDialogCancel>}
                        <AlertDialogAction onClick={onAction}>{actionText}</AlertDialogAction>
                    </AlertDialogFooter>
                ) : null}
            </AlertDialogContent>
        </AlertDialog>
    );
}

export const GlobalAlertDialog = React.memo(View);
