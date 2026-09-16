import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";

import { SpecExportPanel } from "../building-blocks";

export function OperationsExportRoute() {
    return (
        <section className={cn(listBox)}>
            <SpecExportPanel
                scope={{ type: "global" }}
                scopeLabel="this installation"
            />
        </section>
    );
}
