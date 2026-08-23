import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { getModulePermissionDeniedMessage } from "@application/shared/permissions";

export function PermissionReadonlyNotice({ className }: { className?: string } = {}) {
    return (
        <div className={cn(dashedBorderBox, "text-sm leading-6 mb-4", className)}>
            <span className="font-semibold text-orange-500">Notice:</span> {getModulePermissionDeniedMessage("write")}
        </div>
    );
}
