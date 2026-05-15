import React from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditAppConfigFileDialog } from "~/projects/dialogs/create-or-edit-app-config-file/hooks";
import type { AppConfigFile } from "~/projects/domain";

function View({ projectId, appId, configFile }: Props) {
    const { actions: configFileDialogActions } = useCreateOrEditAppConfigFileDialog();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                configFileDialogActions.openEdit(projectId, appId, configFile);
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit app config file</span>
        </Button>
    );
}

interface Props {
    projectId: string;
    appId: string;
    configFile: AppConfigFile;
}

export const EditCell = React.memo(View);
