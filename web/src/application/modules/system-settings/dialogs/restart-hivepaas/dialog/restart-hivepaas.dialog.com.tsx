import { Dialog, DialogDescription, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { Separator } from "@components/ui/separator";
import { toast } from "sonner";
import { HivePaaSRestartCommands } from "~/system-settings/data";

import { RestartHivePaaSForm } from "../form/restart-hivepaas.form.com";
import { useRestartHivePaaSDialogState } from "../hooks";
import type { RestartHivePaaSFormSchemaOutput } from "../schemas";

const fnPlaceholder = () => null;

export function RestartHivePaaSDialog() {
    const { state, props: { onClose = fnPlaceholder } = {}, ...actions } = useRestartHivePaaSDialogState();

    const { mutate: restartHivePaaS, isPending } = HivePaaSRestartCommands.useRestart({
        onSuccess: () => {
            toast.success("Restart request submitted successfully");
            actions.close();
            onClose();
        },
        onError: err => {
            toast.error(err.message || "Failed to restart HivePaaS services");
        },
    });

    function onSubmit(values: RestartHivePaaSFormSchemaOutput) {
        restartHivePaaS({
            payload: values,
        });
    }

    const open = state.mode !== "closed";

    return (
        <Dialog
            open={open}
            onOpenChange={actions.close}
        >
            <DialogFixedContent className="w-full sm:w-[450px] sm:max-w-[450px]">
                <DialogHeader>
                    <DialogTitle>Restart HivePaaS</DialogTitle>
                </DialogHeader>
                <DialogDescription className="sr-only">Restart HivePaaS services</DialogDescription>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <RestartHivePaaSForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                />
            </DialogFixedContent>
        </Dialog>
    );
}
