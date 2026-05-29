import { memo } from "react";

import { Avatar } from "@components/ui";
import { Button } from "@components/ui";
import { format } from "date-fns";
import { BadgeCheck, Check, Clock, KeyRound, Lock, Trash2, User } from "lucide-react";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { UsersCommands } from "~/user-management/data/commands";
import { UsersQueries } from "~/user-management/data/queries";
import { UserRoleBadge, UserStatusBadge } from "~/user-management/module-shared/components";

import { BackButton } from "@application/shared/components";
import { PopConfirm } from "@application/shared/components/pop-confirm";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { EUserStatus } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { useResetUserPasswordDialog } from "@application/modules/user-management/dialogs";

import { UserBreadcrumbs } from "../building-blocks";

import { SingleUserHeaderSkeleton } from "./single-user-header.skeleton.com";

export function View({ userId }: Props) {
    const { data, isLoading, error } = UsersQueries.useFindOneById({ id: userId });
    const { navigate } = useAppNavigate();
    const { canWrite, canDelete } = useConditionalModule({ id: MODULE_IDS.User });
    const { mutate: updateOne, isPending: isUpdating } = UsersCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("User status updated successfully");
        },
    });

    const { mutate: deleteOne, isPending: isDeleting } = UsersCommands.useDeleteOne({});

    const resetUserPasswordDialog = useResetUserPasswordDialog({});

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
        if (!canWrite) {
            return;
        }

        updateOne({ user: { status: EUserStatus.Disabled, id: user.id } });
    };

    const handleActivate = () => {
        if (!canWrite) {
            return;
        }

        updateOne({ user: { status: EUserStatus.Active, id: user.id } });
    };

    const handleRemove = () => {
        if (!canDelete) {
            return;
        }

        deleteOne(
            { id: user.id },
            {
                onSuccess: () => {
                    toast.success("User removed successfully");
                    navigate.modules(ROUTE.userManagement.users.$route);
                },
            },
        );
    };

    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <div className="flex items-center justify-between">
                <UserBreadcrumbs user={user} />
                <div className="flex items-center gap-2">
                    {shouldShowToggleButtons && showDisable && canWrite && (
                        <PopConfirm
                            title="Disable User"
                            variant="destructive"
                            confirmText="Disable"
                            cancelText="Cancel"
                            description="Confirm disabling of this user?"
                            onConfirm={handleDisable}
                        >
                            <Button
                                variant="outline"
                                isLoading={isUpdating}
                            >
                                <Lock className="mr-2 size-4" />
                                Disable
                            </Button>
                        </PopConfirm>
                    )}
                    {shouldShowToggleButtons && showDisable && !canWrite && (
                        <PermissionTooltipAction
                            id={MODULE_IDS.User}
                            action="write"
                        >
                            {({ isDenied }) => (
                                <Button
                                    variant="outline"
                                    disabled={isDenied}
                                >
                                    <Lock className="mr-2 size-4" />
                                    Disable
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}
                    {shouldShowToggleButtons && showActivate && (
                        <PermissionTooltipAction
                            id={MODULE_IDS.User}
                            action="write"
                        >
                            {({ isDenied }) => (
                                <Button
                                    variant="outline"
                                    onClick={handleActivate}
                                    disabled={isDenied}
                                    isLoading={isUpdating}
                                >
                                    <Check className="mr-2 size-4" />
                                    Activate
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}
                    <PermissionTooltipAction
                        id={MODULE_IDS.User}
                        action="write"
                    >
                        {({ isDenied }) => (
                            <Button
                                variant="outline"
                                onClick={() => {
                                    if (!canWrite) {
                                        return;
                                    }

                                    resetUserPasswordDialog.actions.open(user);
                                }}
                                disabled={isDenied}
                            >
                                <KeyRound className="mr-2 size-4" />
                                Reset password
                            </Button>
                        )}
                    </PermissionTooltipAction>
                    {canDelete ? (
                        <PopConfirm
                            title="Remove User"
                            variant="destructive"
                            confirmText="Remove"
                            cancelText="Cancel"
                            description="Confirm deletion of this item?"
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
                    ) : (
                        <PermissionTooltipAction
                            id={MODULE_IDS.User}
                            action="delete"
                        >
                            {({ isDenied }) => (
                                <Button
                                    variant="outline"
                                    disabled={isDenied}
                                >
                                    <Trash2 className="mr-2 size-4" />
                                    Remove
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}
                </div>
            </div>
            <div className="flex items-center gap-4 mt-4 pb-4">
                <BackButton />
                <div className="flex items-center gap-4">
                    <Avatar
                        name={user.fullName}
                        src={user.photo}
                        className="size-20 text-2xl"
                    />
                    <div className="flex flex-col gap-3">
                        <div className="flex items-center gap-2">
                            <h2 className="text-[20px] font-semibold text-foreground">{user.fullName}</h2>
                            <UserStatusBadge status={user.status} />
                        </div>
                        <div className="flex items-center gap-4 text-sm text-muted-foreground">
                            <div className="flex items-center gap-1.5">
                                <User className="size-4 text-blue-500" />
                                <div className="flex gap-1 items-center">
                                    <span>Position:</span>
                                    <span className="text-black">{user.position || "Unknown position"}</span>
                                </div>
                            </div>
                            <span>•</span>
                            <div className="flex gap-1.5 items-center">
                                <BadgeCheck className="size-4 text-blue-500" />
                                <div className="flex gap-1 items-center">
                                    <span>Role:</span>
                                    <UserRoleBadge role={user.role} />
                                </div>
                            </div>
                            {user.lastAccess && (
                                <>
                                    <span>•</span>
                                    <div className="flex items-center gap-1.5">
                                        <Clock className="size-4 text-blue-500" />
                                        <div className="flex gap-1">
                                            <span>Last access:</span>
                                            <span className="text-black">
                                                {format(user.lastAccess, "yyyy-MM-dd HH:mm:ss")}{" "}
                                            </span>
                                        </div>
                                    </div>
                                </>
                            )}
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
