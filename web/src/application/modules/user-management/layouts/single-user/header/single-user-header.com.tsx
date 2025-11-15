import { memo } from "react";

import { Button } from "@components/ui";
import { Check, Lock, Trash2 } from "lucide-react";
import { toast } from "sonner";
import invariant from "tiny-invariant";

import { PopConfirm } from "@application/shared/components/pop-confirm";
import { EUserStatus } from "@application/shared/enums";

import { UsersCommands } from "@application/modules/user-management/data/commands";
import { UsersQueries } from "@application/modules/user-management/data/queries";

import { UserBreadcrumbs } from "../building-blocks";

import { SingleUserHeaderSkeleton } from "./single-user-header.skeleton.com";

export function View({ userId }: Props) {
    const { data, isLoading, error } = UsersQueries.useFindOneById({ id: userId });

    const { mutate: updateOne, isPending: isUpdating } = UsersCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("User status updated successfully");
        },
    });

    const { mutate: deleteOne, isPending: isDeleting } = UsersCommands.useDeleteOne({
        onSuccess: () => {
            toast.success("User removed successfully");
        },
    });

    if (isLoading) {
        return <SingleUserHeaderSkeleton />;
    }

    if (error) {
        return null;
    }

    invariant(data, "data must be defined");

    const { data: user } = data;

    // Logic: If status = 'active' or 'pending': show 'Disable'
    //        If status = 'disabled': show 'Activate'
    //        If status = 'pending': do not show Disable/Activate buttons
    //        Always show: 'Remove'
    const showDisable = user.status === EUserStatus.Active || user.status === EUserStatus.Pending;
    const showActivate = user.status === EUserStatus.Disabled;
    const shouldShowToggleButtons = user.status !== EUserStatus.Pending;

    const handleDisable = () => {
        updateOne({
            id: user.id,
            data: {
                status: EUserStatus.Disabled,
            },
        });
    };

    const handleActivate = () => {
        updateOne({
            id: user.id,
            data: {
                status: EUserStatus.Active,
            },
        });
    };

    const handleRemove = () => {
        deleteOne({ id: user.id });
    };

    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <div className="flex items-center justify-between">
                <UserBreadcrumbs user={user} />
                <div className="flex items-center gap-2">
                    {shouldShowToggleButtons && showDisable && (
                        <Button
                            variant="outline"
                            onClick={handleDisable}
                            isLoading={isUpdating}
                        >
                            <Lock className="mr-2 size-4" />
                            Disable
                        </Button>
                    )}
                    {shouldShowToggleButtons && showActivate && (
                        <Button
                            variant="outline"
                            onClick={handleActivate}
                            isLoading={isUpdating}
                        >
                            <Check className="mr-2 size-4" />
                            Activate
                        </Button>
                    )}
                    <PopConfirm
                        title="Remove user"
                        variant="destructive"
                        confirmText="Remove"
                        cancelText="Cancel"
                        description="Are you sure you want to remove this user?"
                        onConfirm={handleRemove}
                    >
                        <Button
                            variant="outline"
                            disabled={isDeleting}
                        >
                            <Trash2 className="mr-2 size-4" />
                            Remove
                        </Button>
                    </PopConfirm>
                </div>
            </div>
            <div className="py-3 border-b border-border">
                <h1 className="text-2xl font-bold">User</h1>
            </div>
        </div>
    );
}

interface Props {
    userId: string;
}
export const SingleUserHeader = memo(View);
