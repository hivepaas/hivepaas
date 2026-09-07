import { Dialog, DialogDescription, DialogFixedContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { Separator } from "@components/ui/separator";
import { toast } from "sonner";
import { TraefikRestartCommands } from "~/system-settings/data";

import { RestartTraefikForm } from "../form/restart-traefik.form.com";
import { useRestartTraefikDialogState } from "../hooks";
import type { RestartTraefikFormSchemaOutput } from "../schemas";

const fnPlaceholder = () => null;

export function RestartTraefikDialog() {
    const { state, props: { onClose = fnPlaceholder } = {}, ...actions } = useRestartTraefikDialogState();

    const { mutate: restartTraefik, isPending } = TraefikRestartCommands.useRestart({
        onSuccess: () => {
            toast.success("Restart request submitted successfully");
            actions.close();
            onClose();
        },
        onError: err => {
            toast.error(err.message || "Failed to restart Traefik service");
        },
    });

    function onSubmit(_values: RestartTraefikFormSchemaOutput) {
        restartTraefik();
    }

    const open = state.mode !== "closed";

    return (
        <Dialog
            open={open}
            onOpenChange={actions.close}
        >
            <DialogFixedContent className="w-full sm:w-[450px] sm:max-w-[450px]">
                <DialogHeader>
                    <DialogTitle>Restart Traefik</DialogTitle>
                </DialogHeader>
                <DialogDescription className="sr-only">Restart Traefik service</DialogDescription>
                <div className="px-4">
                    <Separator className="opacity-50" />
                </div>
                <RestartTraefikForm
                    isPending={isPending}
                    onSubmit={onSubmit}
                />
            </DialogFixedContent>
        </Dialog>
    );
}
