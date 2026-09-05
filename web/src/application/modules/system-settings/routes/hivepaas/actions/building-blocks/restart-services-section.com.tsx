import { useRestartHivePaaSDialogState } from "~/system-settings/dialogs";
import { ActionExecutePanel } from "~/system-settings/module-shared";

import { MODULE_IDS } from "@application/shared/constants";

export function RestartServicesSection() {
    const restartHivePaaSDialog = useRestartHivePaaSDialogState();

    return (
        <ActionExecutePanel
            message="You can restart HivePaaS services here. Please note that you should only perform this action when really necessary."
            buttonLabel="Restart HivePaaS"
            buttonVariant="destructive"
            isLoading={false}
            permissionModuleId={MODULE_IDS.System}
            onExecute={() => {
                restartHivePaaSDialog.open();
            }}
        />
    );
}
