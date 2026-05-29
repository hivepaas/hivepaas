import { Plus } from "lucide-react";
import type { ProjectSettingsImportKind } from "~/projects/data/commands";
import { useImportProjectSettingsDialog } from "~/projects/dialogs/import-project-settings";

import { Button } from "@/components/ui";

import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction } from "@application/shared/permissions";

export function ProjectSettingsImportButton({ projectId, settingKind }: Props) {
    const importDialog = useImportProjectSettingsDialog();

    return (
        <PermissionTooltipAction
            id={MODULE_IDS.Project}
            action="write"
        >
            {({ isDenied }) => (
                <Button
                    variant="outline"
                    onClick={() => {
                        importDialog.actions.open(projectId, settingKind);
                    }}
                    disabled={isDenied}
                >
                    <Plus className="size-4" />
                    Import
                </Button>
            )}
        </PermissionTooltipAction>
    );
}

interface Props {
    projectId: string;
    settingKind: ProjectSettingsImportKind;
}
