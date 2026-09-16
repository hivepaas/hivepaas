import { SpecExportPanel } from "../building-blocks";

export function OperationsExportRoute() {
    return (
        <div className="flex flex-col gap-6">
            <SpecExportPanel
                scope={{ type: "global" }}
                scopeLabel="this installation"
            />
        </div>
    );
}
