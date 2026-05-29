import { Button } from "@/components/ui";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";

export function ProjectPermissionSubmitButton({ isPending, label = "Save", className = "min-w-[100px]" }: Props) {
    return (
        <PermissionTooltipAction
            id={MODULE_IDS.Project}
            action="write"
        >
            {({ isDenied }) => (
                <Button
                    type="submit"
                    className={className}
                    disabled={isPending || isDenied}
                    isLoading={isPending}
                >
                    {label}
                </Button>
            )}
        </PermissionTooltipAction>
    );
}

interface Props {
    isPending: boolean;
    label?: string;
    className?: string;
}
