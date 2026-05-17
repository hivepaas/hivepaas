import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectCloudStorageCommands } from "~/projects/data/commands";
import { CloudStorageCommands } from "~/settings/data/commands";
import { useUpdateCloudStorageStatusDialog } from "~/settings/dialogs/update-cloud-storage-status";
import type { SettingCloudStorage } from "~/settings/domain";

import { PopConfirm } from "@application/shared/components";

import type { CloudStorageTableScope } from "../cloud-storage-table.types";

function View({ scope, cloudStorage }: Props) {
    const [open, setOpen] = useState(false);
    const updateStatusDialog = useUpdateCloudStorageStatusDialog();

    const { mutate: deleteSettingCloudStorage, isPending: isDeletingSetting } = CloudStorageCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("Cloud storage deleted successfully");
            setOpen(false);
        },
    });

    const { mutate: deleteProjectCloudStorage, isPending: isDeletingProject } =
        ProjectCloudStorageCommands.useDeleteOne({
            onSuccess: () => {
                toast.success("Project cloud storage deleted successfully");
                setOpen(false);
            },
        });

    const isDeleting = isDeletingSetting || isDeletingProject;

    function handleDelete() {
        if (scope.type === "project") {
            deleteProjectCloudStorage({ projectID: scope.projectId, id: cloudStorage.id });
            return;
        }

        deleteSettingCloudStorage({ id: cloudStorage.id });
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
                            updateStatusDialog.actions.open(scope, cloudStorage.id);
                            setOpen(false);
                        }}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </Button>
                    <PopConfirm
                        title="Delete cloud storage"
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
    scope: CloudStorageTableScope;
    cloudStorage: SettingCloudStorage;
}

export const CloudStorageMenuCell = memo(View);
