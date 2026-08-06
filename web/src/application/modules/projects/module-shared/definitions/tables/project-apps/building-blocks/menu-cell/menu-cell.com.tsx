import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { Copy, MoreVertical } from "lucide-react";

import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

function View({ projectId, appId, appEnv }: Props) {
    const [open, setOpen] = useState(false);
    const { navigate } = useAppNavigate();
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    function handleCloneApp() {
        if (!canWrite) {
            return;
        }

        navigate.modules(ROUTE.projects.single.apps.single.configuration.appClone.$route(projectId, appEnv, appId));
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
                        onClick={handleCloneApp}
                    >
                        <Copy className="mr-2 size-4" />
                        Clone App
                    </Button>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    projectId: string;
    appId: string;
    appEnv: string;
}

export const MenuCell = React.memo(View);
