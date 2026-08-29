import { memo } from "react";

import { Avatar, Button } from "@components/ui";
import { format } from "date-fns";
import { BadgeCheck, Check, Clock, KeyRound, Lock, Trash2, User } from "lucide-react";
import { toast } from "sonner";
import invariant from "tiny-invariant";
import { UsersCommands } from "~/user-management/data/commands";
import { UsersQueries } from "~/user-management/data/queries";
import { UserRoleBadge, UserStatusBadge } from "~/user-management/module-shared/components";

import { PopConfirm } from "@application/shared/components/pop-confirm";
import { MODULE_IDS, ROUTE } from "@application/shared/constants";
import { EUserStatus } from "@application/shared/enums";
import { useAppNavigate } from "@application/shared/hooks/router";
import { PermissionTooltipAction, useConditionalModule } from "@application/shared/permissions";

import { useResetUserPasswordDialog } from "@application/modules/user-management/dialogs";

import { Separator } from "@/components/ui/separator";

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
        <div className="bg-background py-3 sm:py-4 px-4 sm:px-5 rounded-lg">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2.5 pb-3">
                <UserBreadcrumbs user={user} />
                <div className="flex items-center gap-1.5 sm:gap-2 flex-wrap">
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
                                size="sm"
                                isLoading={isUpdating}
                                className="h-8 px-2 sm:px-3 text-xs sm:text-sm shrink-0"
                            >
                                <Lock className="hidden sm:inline-block mr-1.5 size-3.5" />
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
                                    size="sm"
                                    disabled={isDenied}
                                    className="h-8 px-2 sm:px-3 text-xs sm:text-sm shrink-0"
                                >
                                    <Lock className="hidden sm:inline-block mr-1.5 size-3.5" />
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
                                    size="sm"
                                    onClick={handleActivate}
                                    disabled={isDenied}
                                    isLoading={isUpdating}
                                    className="h-8 px-2 sm:px-3 text-xs sm:text-sm shrink-0"
                                >
                                    <Check className="hidden sm:inline-block mr-1.5 size-3.5" />
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
                                size="sm"
                                onClick={() => {
                                    if (!canWrite) {
                                        return;
                                    }

                                    resetUserPasswordDialog.actions.open(user);
                                }}
                                disabled={isDenied}
                                className="h-8 px-2 sm:px-3 text-xs sm:text-sm shrink-0"
                            >
                                <KeyRound className="hidden sm:inline-block mr-1.5 size-3.5" />
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
                                size="sm"
                                disabled={isDeleting}
                                className="h-8 px-2 sm:px-3 text-xs sm:text-sm shrink-0"
                            >
                                <Trash2 className="hidden sm:inline-block mr-1.5 size-3.5" />
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
                                    size="sm"
                                    disabled={isDenied}
                                    className="h-8 px-2 sm:px-3 text-xs sm:text-sm shrink-0"
                                >
                                    <Trash2 className="hidden sm:inline-block mr-1.5 size-3.5" />
                                    Remove
                                </Button>
                            )}
                        </PermissionTooltipAction>
                    )}
                </div>
            </div>

            <Separator className="opacity-50" />

            <div className="flex items-center gap-3 sm:gap-4 mt-3 sm:mt-4 pb-1">
                <Avatar
                    name={user.fullName || user.username}
                    src={user.photo}
                    className="size-12 sm:size-16 md:size-20 text-lg sm:text-xl md:text-2xl shrink-0"
                />
                <div className="flex flex-col gap-1.5 sm:gap-2.5 min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                        <h2 className="text-base sm:text-lg font-semibold text-foreground truncate">
                            {user.fullName || user.username}
                        </h2>
                        <UserStatusBadge status={user.status} />
                    </div>
                    <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs sm:text-sm text-muted-foreground">
                        <div className="flex items-center gap-1.5">
                            <User className="size-3.5 sm:size-4 text-blue-500 shrink-0" />
                            <div className="flex gap-1 items-center">
                                <span>Position:</span>
                                <span className="text-foreground font-medium">
                                    {user.position || "Unknown position"}
                                </span>
                            </div>
                        </div>
                        <span className="hidden sm:inline text-amber-500/50 dark:text-amber-400/50 select-none">•</span>
                        <div className="flex gap-1.5 items-center">
                            <BadgeCheck className="size-3.5 sm:size-4 text-blue-500 shrink-0" />
                            <div className="flex gap-1 items-center">
                                <span>Role:</span>
                                <UserRoleBadge role={user.role} />
                            </div>
                        </div>
                        {user.lastAccess && (
                            <>
                                <span className="hidden sm:inline text-amber-500/50 dark:text-amber-400/50 select-none">
                                    •
                                </span>
                                <div className="flex items-center gap-1.5">
                                    <Clock className="size-3.5 sm:size-4 text-blue-500 shrink-0" />
                                    <div className="flex gap-1">
                                        <span>Last access:</span>
                                        <span className="text-foreground font-medium">
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
    );
}

interface Props {
    userId: string;
}
export const SingleUserHeader = memo(View);
