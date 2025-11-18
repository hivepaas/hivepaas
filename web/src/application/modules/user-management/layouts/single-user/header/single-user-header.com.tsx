import { memo } from "react";

import { Avatar } from "@components/ui";
import { Button } from "@components/ui";
import { BadgeCheck, Check, Lock, Trash2, User } from "lucide-react";
import { toast } from "sonner";
import invariant from "tiny-invariant";

import { BackButton } from "@application/shared/components";
import { PopConfirm } from "@application/shared/components/pop-confirm";
import { EUserRole, EUserStatus } from "@application/shared/enums";

import { UsersCommands } from "@application/modules/user-management/data/commands";
import { UsersQueries } from "@application/modules/user-management/data/queries";
import { UserStatusBadge } from "@application/modules/user-management/module-shared/components";

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

    const roleMap: Record<EUserRole, string> = {
        [EUserRole.Admin]: "Admin",
        [EUserRole.Member]: "Member",
    };

    const handleDisable = () => {
        console.log("disable");
    };

    const handleActivate = () => {
        console.log("activate");
    };

    const handleRemove = () => {
        console.log("remove");
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
            <div className="flex items-center gap-4 mt-4 pb-4">
                <BackButton />
                <div className="flex items-center gap-4">
                    <Avatar
                        name={user.fullName}
                        src={user.photo}
                        className="size-20"
                    />
                    <div className="flex flex-col gap-3">
                        <div className="flex items-center gap-2">
                            <h2 className="text-[20px] font-semibold text-foreground">{user.fullName}</h2>
                            <UserStatusBadge status={user.status} />
                        </div>
                        <div className="flex items-center gap-4 text-sm text-muted-foreground">
                            <div className="flex items-center gap-1.5">
                                <User className="size-4 text-blue-500" />
                                <span>{user.position}</span>
                            </div>
                            <span>•</span>
                            <div className="flex items-center gap-1.5">
                                <BadgeCheck className="size-4 text-blue-500" />
                                <span>Role: {roleMap[user.role] || user.role}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
}

interface Props {
    userId: string;
}
export const SingleUserHeader = memo(View);
