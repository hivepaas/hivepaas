import { useRestartTraefikDialogState } from "~/system-settings/dialogs";
import { ActionExecutePanel } from "~/system-settings/module-shared";

import { MODULE_IDS } from "@application/shared/constants";

export function RestartTraefikSection() {
    const restartTraefikDialog = useRestartTraefikDialogState();

    return (
        <ActionExecutePanel
            message="You can restart Traefik service here. Please note that you should only perform this action when really necessary."
            buttonLabel="Restart Traefik"
            buttonVariant="destructive"
            isLoading={false}
            permissionModuleId={MODULE_IDS.System}
            onExecute={() => {
                restartTraefikDialog.open();
            }}
        />
    );
}
