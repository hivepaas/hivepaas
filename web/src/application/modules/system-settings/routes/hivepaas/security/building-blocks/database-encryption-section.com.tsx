import { useChangeKekDialogState } from "~/system-settings/dialogs";
import { SectionHeader } from "~/system-settings/module-shared";
import { ActionExecutePanel } from "~/system-settings/module-shared";

import { MODULE_IDS } from "@application/shared/constants";

export function DatabaseEncryptionSection() {
    const changeKekDialog = useChangeKekDialogState();

    return (
        <div className="flex flex-col gap-3">
            <SectionHeader>Database Encryption</SectionHeader>
            <div className="px-3">
                <ActionExecutePanel
                    message="All your sensitive data is encrypted in the database. HivePaaS utilizes a Key Encryption Key (KEK) and Data Encryption Key (DEK) architecture. The DEK is encrypted and stored in the database, while the KEK is used to decrypt the DEK and is never stored in the DB. You must ensure you can provide the KEK when needed, so please store your KEK in a safe and secure place."
                    buttonLabel="Change App Secret (KEK)"
                    isLoading={false}
                    permissionModuleId={MODULE_IDS.System}
                    onExecute={() => {
                        changeKekDialog.open();
                    }}
                />
            </div>
        </div>
    );
}
