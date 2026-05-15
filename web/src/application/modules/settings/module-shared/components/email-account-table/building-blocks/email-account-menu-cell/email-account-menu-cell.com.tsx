import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { Edit, MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectEmailCommands } from "~/projects/data/commands";
import { EmailCommands } from "~/settings/data/commands";
import { useCreateOrEditEmailAccountDialog } from "~/settings/dialogs/create-or-edit-email-account";
import { useUpdateEmailAccountStatusDialog } from "~/settings/dialogs/update-email-account-status";
import type { SettingEmail } from "~/settings/domain";

import { PopConfirm } from "@application/shared/components";

import type { EmailAccountTableScope } from "../../email-account-table.types";

function View({ scope, emailAccount }: Props) {
    const [open, setOpen] = useState(false);

    const createOrEditDialog = useCreateOrEditEmailAccountDialog();
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
                        onClick={() => {
                            createOrEditDialog.actions.openEdit(scope, emailAccount.id);
                            setOpen(false);
                        }}
                    >
                        <Edit className="mr-2 size-4" />
                        Edit
                    </Button>
                    <Button
                        className="justify-start py-1.5"
                        variant="ghost"
                        onClick={() => {
                            updateStatusDialog.actions.open(scope, emailAccount.id);
                            setOpen(false);
                        }}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </Button>
                    <PopConfirm
                        title="Delete Email account"
                        variant="destructive"
                        confirmText="Delete"
                        cancelText="Cancel"
                        description="Confirm deletion of this item?"
                        onConfirm={handleDelete}
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
