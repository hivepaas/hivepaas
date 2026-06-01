import { Button } from "@/components/ui";

import { ProjectPermissionTooltipAction } from "../project-permission-tooltip-action";

export function ProjectPermissionSubmitButton({
    projectId,
    isPending,
    label = "Save",
    className = "min-w-[100px]",
}: Props) {
    return (
        <ProjectPermissionTooltipAction
            projectId={projectId}
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
        </ProjectPermissionTooltipAction>
    );
}

interface Props {
    projectId?: string;
    isPending: boolean;
    label?: string;
    className?: string;
}
