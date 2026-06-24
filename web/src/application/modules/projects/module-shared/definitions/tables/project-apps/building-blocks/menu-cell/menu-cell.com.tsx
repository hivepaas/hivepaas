import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { Copy, MoreVertical } from "lucide-react";
import { useCopyProjectAppDialog } from "~/projects/dialogs/copy-project-app";

import { MODULE_IDS } from "@application/shared/constants";
import { useConditionalModule } from "@application/shared/permissions";

function View({ projectId, appId }: Props) {
    const [open, setOpen] = useState(false);
    const copyProjectAppDialog = useCopyProjectAppDialog();
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    function handleCopyApp() {
        if (!canWrite) {
            return;
        }

        copyProjectAppDialog.actions.open(projectId, appId);
        setOpen(false);
    }

    return (
        <DropdownMenu
            open={open}
            onOpenChange={setOpen}
        >
            <DropdownMenuTrigger asChild>
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                >
                    <MoreVertical className="size-4" />
                    <span className="sr-only">Actions menu</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <div className="flex flex-col gap-0">
                    <Button
                        className="justify-start py-1.5"
                        variant="ghost"
                        disabled={!canWrite}
                        onClick={handleCopyApp}
                    >
                        <Copy className="mr-2 size-4" />
                        Copy App
                    </Button>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    projectId: string;
    appId: string;
}

export const MenuCell = React.memo(View);
