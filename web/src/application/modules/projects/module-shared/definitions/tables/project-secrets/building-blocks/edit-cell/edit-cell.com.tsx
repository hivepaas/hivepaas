import React from "react";

import { Button } from "@components/ui/button";
import { EyeIcon } from "lucide-react";
import { useCreateOrEditProjectSecretDialog } from "~/projects/dialogs/create-or-edit-project-secret/hooks";
import type { ProjectSecret } from "~/projects/domain";

function View({ projectId, secret }: Props) {
    const { actions: secretDialogActions } = useCreateOrEditProjectSecretDialog();

    return (
        <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-link hover:opacity-50"
            onClick={() => {
                secretDialogActions.openEdit(projectId, secret);
            }}
        >
            <EyeIcon className="size-5" />
            <span className="sr-only">Edit project secret</span>
        </Button>
    );
}

interface Props {
    projectId: string;
    secret: ProjectSecret;
}

export const EditCell = React.memo(View);
