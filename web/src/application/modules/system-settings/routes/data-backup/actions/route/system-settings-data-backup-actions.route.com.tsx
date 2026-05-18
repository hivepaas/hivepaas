import { toast } from "sonner";
import { SystemBackupCommands } from "~/system-settings/data";
import { ActionExecutePanel } from "~/system-settings/module-shared";

export function SystemSettingsDataBackupActionsRoute() {
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
            onExecute={() => {
                execute({});
            }}
        />
    );
}
