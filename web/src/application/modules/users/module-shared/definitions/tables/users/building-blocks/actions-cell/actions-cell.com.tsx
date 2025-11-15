import React from "react";

import { EyeIcon, Trash2Icon } from "lucide-react";
import { toast } from "sonner";

import { AppLink } from "@application/shared/components/navigation";
import { PopConfirm } from "@application/shared/components/pop-confirm";
import { ROUTE } from "@application/shared/constants";

import { UsersCommands } from "@application/modules/users/data/commands";

function View({ id }: Props) {
    const { mutate: deleteOne } = UsersCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("User deleted successfully");
        },
    });
    return (
        <div className="flex items-center justify-center gap-4">
            <AppLink.Modules
                className="flex items-center justify-center text-link hover:opacity-50 transition-opacity duration-200"
                to={ROUTE.userManagement.users.details.$route(id)}
            >
                <EyeIcon className="size-5" />
            </AppLink.Modules>

            <PopConfirm
                title="Remove user"
                variant="destructive"
                confirmText="Remove"
                cancelText="Cancel"
                description="Are you sure you want to remove this user?"
                onConfirm={() => {
                    deleteOne({ id });
                }}
            >
                <Trash2Icon className="size-5 text-destructive hover:opacity-50 transition-opacity duration-200" />
            </PopConfirm>
        </div>
    );
}

interface Props {
    id: string;
}

export const ActionsCell = React.memo(View);
