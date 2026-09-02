import { toast } from "sonner";
import { SystemBackupRepoCleanupCommands } from "~/system-settings/data";
import { ActionExecutePanel } from "~/system-settings/module-shared";

import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

export function SystemSettingsBackupRepoCleanupActionsRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { mutate: execute, isPending } = SystemBackupRepoCleanupCommands.useExecute({
        onSuccess: () => {
            toast.success("Cleanup started");
        },
    });

    return (
        <ActionExecutePanel
            message="Make sure you have enabled the backup repo cleanup job before performing this action."
            buttonLabel="Run Cleanup Now"
            isLoading={isPending}
            permissionModuleId={MODULE_IDS.System}
            onExecute={() => {
                if (!canWrite) {
                    return;
                }

                execute({ targetRepos: [] });
            }}
        />
    );
}
