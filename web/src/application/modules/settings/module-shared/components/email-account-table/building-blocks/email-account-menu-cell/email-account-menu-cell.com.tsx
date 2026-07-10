import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectEmailCommands } from "~/projects/data/commands";
import { EmailCommands } from "~/settings/data/commands";
import { useUpdateEmailAccountStatusDialog } from "~/settings/dialogs/update-email-account-status";
import type { SettingEmail } from "~/settings/domain";
import { SettingsScopeMenuButton, SettingsScopePopConfirmButton } from "~/settings/module-shared/components";
import { SETTINGS_ENTITY_TITLES } from "~/settings/module-shared/constants/settings-entity-titles";
import { isInheritedProjectSetting } from "~/settings/module-shared/hooks";

import type { EmailAccountTableScope } from "../../email-account-table.types";

function View({ scope, emailAccount }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateEmailAccountStatusDialog();
    const { mutate: deleteSettingEmailAccount, isPending: isDeletingSetting } = EmailCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Email account deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectEmailAccount, isPending: isDeletingProject } = ProjectEmailCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project Email account deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSetting || isDeletingProject;
    const isInheritedProject = isInheritedProjectSetting(scope, emailAccount.inherited);

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectEmailAccount({
                projectID: scope.projectId,
                id: emailAccount.id,
            });
            return;
        }

        deleteSettingEmailAccount({ id: emailAccount.id });
    }

    function handleChangeStatus() {
        if (isInheritedProject) {
            updateStatusDialog.actions.open(scope, emailAccount.id, {
                props: {
                    readOnlyInherited: true,
                    entityTitle: SETTINGS_ENTITY_TITLES.emailAccount,
                },
            });
            setOpen(false);
            return;
        }

        updateStatusDialog.actions.open(scope, emailAccount.id);
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
                        onClick={handleChangeStatus}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </SettingsScopeMenuButton>
                    <SettingsScopePopConfirmButton
                        scope={scope}
                        action="delete"
                        title="Delete Email account"
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
    scope: EmailAccountTableScope;
    emailAccount: SettingEmail;
}

export const EmailAccountMenuCell = memo(View);
