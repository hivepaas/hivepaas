import { toast } from "sonner";
import { SystemImageScanCommands } from "~/system-settings/data";
import { ActionExecutePanel } from "~/system-settings/module-shared";

import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

export function SystemSettingsImageScanningActionsRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { mutate: execute, isPending } = SystemImageScanCommands.useExecute({
        onSuccess: () => {
            toast.success("Image scan started");
        },
    });

    return (
        <ActionExecutePanel
            message="Runs the daily scan now: every app's images are scanned again, except those scanned within the hour, and what is new is notified. Image scanning must be switched on."
            buttonLabel="Run Scan Now"
            isLoading={isPending}
            permissionModuleId={MODULE_IDS.System}
            onExecute={() => {
                if (canWrite) {
                    execute();
                }
            }}
        />
    );
}
