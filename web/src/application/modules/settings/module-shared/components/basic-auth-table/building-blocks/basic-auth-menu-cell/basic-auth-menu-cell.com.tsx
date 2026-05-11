import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { Edit, MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectBasicAuthCommands } from "~/projects/data/commands";
import { BasicAuthCommands } from "~/settings/data/commands";
import { useCreateOrEditBasicAuthDialog } from "~/settings/dialogs/create-or-edit-basic-auth";
import { useUpdateBasicAuthStatusDialog } from "~/settings/dialogs/update-basic-auth-status";
import type { SettingBasicAuth } from "~/settings/domain";

import { PopConfirm } from "@application/shared/components";

import type { BasicAuthTableScope } from "../../basic-auth-table.types";

function View({ scope, basicAuth }: Props) {
    const [open, setOpen] = useState(false);

    const createOrEditDialog = useCreateOrEditBasicAuthDialog();
    const updateStatusDialog = useUpdateBasicAuthStatusDialog();

    const { mutate: deleteSettingBasicAuth, isPending: isDeletingSetting } = BasicAuthCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Basic auth deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectBasicAuth, isPending: isDeletingProject } = ProjectBasicAuthCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Project basic auth deleted successfully");
            setOpen(false);
        },
    });

    const isDeleting = isDeletingSetting || isDeletingProject;

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectBasicAuth({
                projectID: scope.projectId,
                id: basicAuth.id,
            });
            return;
        }

        deleteSettingBasicAuth({ id: basicAuth.id });
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
                            createOrEditDialog.actions.openEdit(scope, basicAuth);
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
                            updateStatusDialog.actions.open(scope, basicAuth);
                            setOpen(false);
                        }}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </Button>
                    <PopConfirm
                        title="Delete basic auth"
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
    scope: BasicAuthTableScope;
    basicAuth: SettingBasicAuth;
}

export const BasicAuthMenuCell = memo(View);
