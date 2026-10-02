import { toast } from "sonner";
import { SystemRegistryAuthRenewalCommands } from "~/system-settings/data";
import { ActionExecutePanel } from "~/system-settings/module-shared";

import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

export function SettingsRegistryAuthRenewalActionsRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { mutate: execute, isPending } = SystemRegistryAuthRenewalCommands.useExecute({
        onSuccess: () => {
            toast.success("Renewal started");
        },
    });

    return (
        <ActionExecutePanel
            message="Gets a new token for every Amazon ECR credential now, and hands it to the services that pull with it. Make sure the renewal is enabled before running it."
            buttonLabel="Run Renewal Now"
            isLoading={isPending}
            permissionModuleId={MODULE_IDS.System}
            onExecute={() => {
                if (!canWrite) {
                    return;
                }

                execute({ targetAuths: [] });
            }}
        />
    );
}
