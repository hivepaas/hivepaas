import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, RotateCcw, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectSslCertCommands } from "~/projects/data/commands";
import { SslCertCommands } from "~/settings/data/commands";
import { useUpdateSslCertStatusDialog } from "~/settings/dialogs/update-ssl-cert-status";
import type { SettingSslCert } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { SslCertTableScope } from "../../ssl-cert-table.types";

function View({ scope, sslCert }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateSslCertStatusDialog();
    const { mutate: renewSettingSslCert, isPending: isRenewingSetting } = SslCertCommands.useRenewOne({
        onSuccess: () => {
            toast.success("SSL certificate renewal started");
            setOpen(false);
        },
    });

    const { mutate: renewProjectSslCert, isPending: isRenewingProject } = ProjectSslCertCommands.useRenewOne({
        onSuccess: () => {
            toast.success("SSL certificate renewal started");
            setOpen(false);
        },
    });

    const { mutate: deleteSettingSslCert, isPending: isDeletingSetting } = SslCertCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("SSL certificate deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectSslCert, isPending: isDeletingProject } = ProjectSslCertCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project SSL certificate deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isRenewing = isRenewingSetting || isRenewingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, sslCert.inherited);

    function handleRenew() {
        if (scope.type === "project") {
            renewProjectSslCert({
                projectID: scope.projectId,
                id: sslCert.id,
            });
            return;
        }

        renewSettingSslCert({ id: sslCert.id });
    }

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectSslCert({
                projectID: scope.projectId,
                id: sslCert.id,
            });
            return;
        }

        deleteSettingSslCert({ id: sslCert.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, sslCert.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.sslCert,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, sslCert.id);
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
                    <SettingsScopeMenuButton
                        scope={scope}
                        action="write"
                        isLoading={isRenewing}
                        onClick={handleRenew}
                    >
                        <RotateCcw className="mr-2 size-4" />
                        Renew Now
                    </SettingsScopeMenuButton>
                    <SettingsScopeMenuButton
                        scope={scope}
                        action="write"
                        onClick={handleChangeStatus}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </SettingsScopeMenuButton>
                    <SettingsScopePopConfirmButton
                        scope={scope}
                        action="delete"
                        title="Delete SSL certificate"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Confirm deletion of this item?"
                        onConfirm={handleDelete}
                        isLoading={isDeleting}
                    >
                        <Trash2Icon className="mr-2 size-4" />
                        Delete
                    </SettingsScopePopConfirmButton>
                </div>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

interface Props {
    scope: SslCertTableScope;
    sslCert: SettingSslCert;
}

export const SslCertMenuCell = memo(View);
