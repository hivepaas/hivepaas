import { toast } from "sonner";
import { SystemCleanupCommands } from "~/system-settings/data";
import { ActionExecutePanel } from "~/system-settings/module-shared";

export function SystemSettingsDataCleanupActionsRoute() {
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
            onExecute={() => {
                execute({});
            }}
        />
    );
}
