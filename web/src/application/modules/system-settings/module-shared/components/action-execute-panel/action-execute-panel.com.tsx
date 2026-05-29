import { Button } from "@/components/ui";

import { PermissionTooltipAction } from "@application/shared/permissions";
import type { ModuleId } from "@application/shared/permissions";

export function ActionExecutePanel({ message, buttonLabel, isLoading, permissionModuleId, onExecute }: Props) {
    const button = ({ isDenied = false }: { isDenied?: boolean } = {}) => (
        <Button
            type="button"
            className="min-w-[180px]"
            disabled={isLoading || isDenied}
            isLoading={isLoading}
            onClick={() => {
                if (isDenied) {
                    return;
                }

                onExecute();
            }}
        >
            {buttonLabel}
        </Button>
    );

    return (
        <div className="rounded-lg border bg-background p-6">
            <div className="flex flex-col items-start gap-6">
                <p className="text-base text-foreground">{message}</p>
                {permissionModuleId ? (
                    <PermissionTooltipAction
                        id={permissionModuleId}
                        action="write"
                    >
                        {button}
                    </PermissionTooltipAction>
                ) : (
                    button()
                )}
            </div>
        </div>
    );
}

interface Props {
    message: string;
    buttonLabel: string;
    isLoading: boolean;
    permissionModuleId?: ModuleId;
    onExecute: () => void;
}
