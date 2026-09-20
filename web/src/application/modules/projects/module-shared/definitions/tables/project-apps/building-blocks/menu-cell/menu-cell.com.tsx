import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { Copy, MoreVertical, Trash2 } from "lucide-react";
import type { ProjectAppBaseRef } from "~/projects/domain";

import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { useAppNavigate } from "@application/shared/hooks/router";
import { useConditionalModule } from "@application/shared/permissions";

function View({ projectId, appId, appEnv, parentApp, hideClone = false }: Props) {
    const [open, setOpen] = useState(false);
    const { navigate } = useAppNavigate();
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.Project });

    const hasParentApp = Boolean(parentApp && parentApp.id.trim() !== "");
    const canShowClone = !hideClone && !hasParentApp;

    function handleCloneApp() {
        if (!canWrite) {
            return;
        }

        navigate.modules(ROUTE.projects.single.apps.single.configuration.appClone.$route(projectId, appEnv, appId));
        setOpen(false);
    }

    function handleDeleteApp() {
        if (!canWrite) {
            return;
        }

        navigate.modules(ROUTE.projects.single.apps.single.configuration.dangerZone.$route(projectId, appEnv, appId));
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
                    onClick={e => {
                        e.stopPropagation();
                    }}
                >
                    <MoreVertical className="size-4" />
                    <span className="sr-only">Actions menu</span>
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <div className="flex flex-col gap-0">
                    {canShowClone && (
                        <Button
                            className="justify-start py-1.5"
                            variant="ghost"
                            disabled={!canWrite}
                            onClick={handleCloneApp}
                        >
                            <Copy className="mr-2 size-4" />
                            Clone App
                        </Button>
                    )}
                    <Button
                        className="justify-start py-1.5 text-destructive hover:text-destructive hover:bg-destructive/10"
                        variant="ghost"
                        disabled={!canWrite}
                        onClick={handleDeleteApp}
                    >
                        <Trash2 className="mr-2 size-4" />
                        Delete App
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
    parentApp?: ProjectAppBaseRef | null;
    hideClone?: boolean;
}

export const MenuCell = React.memo(View);
