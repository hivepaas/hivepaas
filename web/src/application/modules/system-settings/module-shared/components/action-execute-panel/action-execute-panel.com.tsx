import { PermissionTooltipAction } from "@application/shared/permissions";
import type { ModuleId } from "@application/shared/permissions";

import { Button } from "@/components/ui";

export function ActionExecutePanel({
    message,
    buttonLabel,
    buttonVariant = "default",
    isLoading,
    permissionModuleId,
    onExecute,
}: Props) {
    const button = ({ isDenied = false }: { isDenied?: boolean } = {}) => (
        <Button
            type="button"
            variant={buttonVariant}
            className="min-w-[120px]"
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
        <div className="rounded-lg border bg-background p-4 ">
            <div className="flex flex-col items-start gap-6">
                <p className="text-sm font-medium text-foreground">{message}</p>
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
    buttonVariant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link";
    isLoading: boolean;
    permissionModuleId?: ModuleId;
    onExecute: () => void;
}
