import { toast } from "sonner";
import { SystemBackupCommands } from "~/system-settings/data";
import { ActionExecutePanel } from "~/system-settings/module-shared";

import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

export function SystemSettingsDataBackupActionsRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { mutate: execute, isPending } = SystemBackupCommands.useExecute({
        onSuccess: () => {
            toast.success("Backup started");
        },
    });

    return (
        <ActionExecutePanel
            message="Make sure you have enabled the backup job before performing this action."
            buttonLabel="Run Backup Now"
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
