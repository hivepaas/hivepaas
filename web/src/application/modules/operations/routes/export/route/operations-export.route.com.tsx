import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";
import type { SpecExportScope, SpecImportScope } from "~/operations/domain";

import { MODULE_IDS } from "@application/shared/constants";

import { SpecExportPanel, SpecImportPanel } from "../building-blocks";

// Constants, so the import panel sees the same scope on every render and asks
// for its plan only when something the plan answers changes.
const EXPORT_SCOPE: SpecExportScope = { type: "global" };
const IMPORT_SCOPE: SpecImportScope = { type: "global" };

export function OperationsExportRoute() {
    return (
        <section className={cn(listBox, "flex flex-col gap-6")}>
            <div className="flex flex-col gap-3">
                <h2 className="text-base font-semibold text-foreground">Export</h2>
                <SpecExportPanel
                    scope={EXPORT_SCOPE}
                    scopeLabel="this installation"
                />
            </div>
            <div className="flex flex-col gap-3">
                <h2 className="text-base font-semibold text-foreground">Import</h2>
                <SpecImportPanel
                    scope={IMPORT_SCOPE}
                    scopeLabel="this installation"
                    permissionModuleId={MODULE_IDS.System}
                />
            </div>
        </section>
    );
}
