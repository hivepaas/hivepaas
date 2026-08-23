import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { inheritedSettingReadonlyDescription } from "~/settings/module-shared/hooks";

export function InheritedSettingReadonlyNotice({ className }: { className?: string } = {}) {
    return (
        <div className={cn(dashedBorderBox, "text-sm leading-6 mb-4", className)}>
            {inheritedSettingReadonlyDescription}
        </div>
    );
}
