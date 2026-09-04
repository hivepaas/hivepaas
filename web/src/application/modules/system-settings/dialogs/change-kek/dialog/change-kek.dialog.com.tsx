import { useState } from "react";

import { Dialog, DialogDescription, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { Separator } from "@components/ui/separator";
import { toast } from "sonner";
import { HivePaaSAppSecretCommands } from "~/system-settings/data";

import { ChangeKekForm } from "../form/change-kek.form.com";
import { useChangeKekDialogState } from "../hooks";
import type { ChangeKekFormSchemaOutput } from "../schemas";

const fnPlaceholder = () => null;

export function ChangeKekDialog() {
    const [hasChanges, setHasChanges] = useState(false);
    const { state, props: { onClose = fnPlaceholder } = {}, ...actions } = useChangeKekDialogState();

    const { mutate: updateAppSecret, isPending } = HivePaaSAppSecretCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("KEK updated successfully");
            actions.close();
            onClose();
        },
        onError: err => {
            toast.error(err.message || "Failed to update KEK");
        },
    });

    function onSubmit(values: ChangeKekFormSchemaOutput) {
        updateAppSecret({
            payload: {
                currentSecret: values.currentSecret.trim() || undefined,
                newSecret: values.newSecret,
            },
        });
    }

    function handleClose() {
        if (hasChanges) {
            const canClose = window.confirm("Are you sure you want to close modal without saving changes?");
            if (!canClose) return;
        }
        actions.close();
    }

    const open = state.mode !== "closed";

    return (
        <Dialog
            open={open}
            onOpenChange={handleClose}
        >
            <DialogFixedContent className="min-w-[400px] w-fit">
                <DialogHeader>
                    <DialogTitle>Change KEK</DialogTitle>
                </DialogHeader>
                <DialogDescription className="sr-only">Change Key Encryption Key</DialogDescription>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <ChangeKekForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                    onHasChanges={setHasChanges}
                />
            </DialogFixedContent>
        </Dialog>
    );
}
