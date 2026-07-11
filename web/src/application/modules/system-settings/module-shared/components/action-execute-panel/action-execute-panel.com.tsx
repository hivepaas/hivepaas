import { FormActionBar } from "@application/shared/components";
import { PermissionTooltipAction } from "@application/shared/permissions";
import type { ModuleId } from "@application/shared/permissions";

import { Button } from "@/components/ui";

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
        <div className="rounded-lg border bg-background">
            <div className="p-6">
                <p className="text-base text-foreground">{message}</p>
            </div>
            <FormActionBar>
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
            </FormActionBar>
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
