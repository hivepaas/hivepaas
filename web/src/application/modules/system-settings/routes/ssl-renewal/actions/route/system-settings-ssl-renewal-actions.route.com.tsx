import { toast } from "sonner";
import { SystemSslRenewalCommands } from "~/system-settings/data";
import { ActionExecutePanel } from "~/system-settings/module-shared";

import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

export function SystemSettingsSslRenewalActionsRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { mutate: execute, isPending } = SystemSslRenewalCommands.useExecute({
        onSuccess: () => {
            toast.success("Renewal started");
        },
    });

    return (
        <ActionExecutePanel
            message="Make sure you have enabled the Certificate renewal job before performing this action."
            buttonLabel="Run Renewal Now"
            isLoading={isPending}
            permissionModuleId={MODULE_IDS.System}
            onExecute={() => {
                if (!canWrite) {
                    return;
                }

                execute({ targetSSLs: [] });
            }}
        />
    );
}
