import { inheritedSettingReadonlyDescription } from "~/settings/module-shared/hooks";

export function InheritedSettingReadonlyNotice() {
    return (
        <div className="border border-dashed border-primary rounded-md bg-muted/40 px-5 py-4 text-center text-sm mb-3">
            {inheritedSettingReadonlyDescription}
        </div>
    );
}
