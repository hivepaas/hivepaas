import React, { useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, PowerIcon, PowerOffIcon, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { AppSettingMountsCommands } from "~/projects/data/commands";
import type { AppSettingMount } from "~/projects/domain";
import { EProjectSecretStatus } from "~/projects/module-shared/enums";

import { PopConfirm } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

function View({ projectId, env, appId, settingMount }: Props) {
    const [open, setOpen] = useState(false);
    const { canDelete } = useConditionalModule({ id: MODULE_IDS.Project });
    const isActive = settingMount.status === EProjectSecretStatus.Active;

    const { mutate: deleteOne, isPending: isDeleting } = AppSettingMountsCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Setting mount deleted");
            setOpen(false);
        },
    });

    // Enabling an entry with a private key goes through the Reveal Secrets gate;
    // a refusal comes back as an error toast.
    const { mutate: updateStatus, isPending: isUpdatingStatus } = AppSettingMountsCommands.useUpdateStatus({
        onSuccess: () => {
            toast.success(isActive ? "Setting mount disabled" : "Setting mount enabled");
            setOpen(false);
        },
    });

    return (
        <DropdownMenu
            open={open}
            onOpenChange={setOpen}
        >
            <DropdownMenuTrigger
                asChild
                className="h-8 w-8"
            >
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
                    <PermissionTooltipAction
                        id={MODULE_IDS.Project}
                        action="write"
                        triggerClassName="w-full"
                    >
                        {({ isDenied }) => (
                            <Button
                                className="justify-start py-1.5 w-full"
                                variant="ghost"
                                disabled={isDenied || isUpdatingStatus}
                                onClick={() => {
                                    updateStatus({
                                        projectID: projectId,
                                        env,
                                        appID: appId,
                                        settingMountID: settingMount.id,
                                        updateVer: settingMount.updateVer,
                                        status: isActive ? EProjectSecretStatus.Disabled : EProjectSecretStatus.Active,
                                    });
                                }}
                            >
                                {isActive ? (
                                    <PowerOffIcon className="mr-2 size-4" />
                                ) : (
                                    <PowerIcon className="mr-2 size-4" />
                                )}
                                {isActive ? "Disable" : "Enable"}
                            </Button>
                        )}
                    </PermissionTooltipAction>
                    {canDelete ? (
                        <PopConfirm
                            title="Delete Item"
                            variant="destructive"
                            confirmText="Delete"
                            cancelText="Cancel"
                            description="Confirm deletion of this item?"
                            onConfirm={() => {
                                deleteOne({ projectID: projectId, env, appID: appId, settingMountID: settingMount.id });
                            }}
                        >
                            <Button
                                className="justify-start py-1.5"
                                variant="ghost"
                                disabled={isDeleting}
                            >
                                <Trash2Icon className="mr-2 size-4" />
                                Remove
                            </Button>
                        </PopConfirm>
                    ) : (
                        <PermissionTooltipAction
                            id={MODULE_IDS.Project}
                            action="delete"
                            triggerClassName="w-full"
                        >
                            {({ isDenied }) => (
                                <Button
                                    className="justify-start py-1.5 w-full"
                                    variant="ghost"
                                    disabled={isDenied}
                                >
                                    <Trash2Icon className="mr-2 size-4" />
                                    Remove
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    settingMount: AppSettingMount;
}

export const MenuCell = React.memo(View);
