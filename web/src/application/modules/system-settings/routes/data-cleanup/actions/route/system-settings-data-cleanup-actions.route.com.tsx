import { toast } from "sonner";
import { SystemCleanupCommands } from "~/system-settings/data";
import { ActionExecutePanel } from "~/system-settings/module-shared";

import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

export function SystemSettingsDataCleanupActionsRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { mutate: execute, isPending } = SystemCleanupCommands.useExecute({
        onSuccess: () => {
            toast.success("Cleanup started");
        },
    });

    return (
        <ActionExecutePanel
            message="Make sure you have enabled the cleanup job before performing this action."
            buttonLabel="Run Cleanup Now"
            isLoading={isPending}
            permissionModuleId={MODULE_IDS.System}
            onExecute={() => {
                if (!canWrite) {
                    return;
                }

                execute({});
            }}
        />
    );
}
