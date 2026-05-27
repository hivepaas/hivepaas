import { Plus } from "lucide-react";
import type { ProjectSettingsImportKind } from "~/projects/data/commands";
import { useImportProjectSettingsDialog } from "~/projects/dialogs/import-project-settings";

import { Button } from "@/components/ui";

export function ProjectSettingsImportButton({ projectId, settingKind }: Props) {
    const importDialog = useImportProjectSettingsDialog();

    return (
        <Button
            variant="outline"
            onClick={() => {
                importDialog.actions.open(projectId, settingKind);
            }}
        >
            <Plus className="size-4" />
            Import
        </Button>
    );
}

interface Props {
    projectId: string;
    settingKind: ProjectSettingsImportKind;
}
