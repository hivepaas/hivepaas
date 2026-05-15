import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuTrigger } from "@components/ui/dropdown-menu";
import { MoreVertical, SlidersHorizontal, Trash2Icon } from "lucide-react";
import { toast } from "sonner";
import { ProjectSslCertCommands } from "~/projects/data/commands";
import { SslCertCommands } from "~/settings/data/commands";
import { useUpdateSslCertStatusDialog } from "~/settings/dialogs/update-ssl-cert-status";
import type { SettingSslCert } from "~/settings/domain";

import { PopConfirm } from "@application/shared/components";

import type { SslCertTableScope } from "../../ssl-cert-table.types";

function View({ scope, sslCert }: Props) {
    const [open, setOpen] = useState(false);

    const updateStatusDialog = useUpdateSslCertStatusDialog();

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
                            updateStatusDialog.actions.open(scope, sslCert.id);
                            setOpen(false);
                        }}
                    >
                        <SlidersHorizontal className="mr-2 size-4" />
                        Change Status
                    </Button>
                    <PopConfirm
                        title="Delete SSL certificate"
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
    scope: SslCertTableScope;
    sslCert: SettingSslCert;
}

export const SslCertMenuCell = memo(View);
