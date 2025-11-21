import { memo } from "react";

import { Avatar } from "@components/ui";
import { format } from "date-fns";
import { BadgeCheck, Clock, User } from "lucide-react";
import invariant from "tiny-invariant";

import { TabNavigation } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { EUserRole } from "@application/shared/enums";

import { UsersQueries } from "@application/modules/user-management/data/queries";
import { UserStatusBadge } from "@application/modules/user-management/module-shared/components";

import { ProfileHeaderSkeleton } from "./profile-header.skeleton.com";

function View({ userId }: Props) {
    const { data, isLoading, error } = UsersQueries.useFindOneById({ id: userId });

    if (isLoading) {
        return <ProfileHeaderSkeleton />;
    }

    if (error) {
        return null;
    }

    invariant(data, "data must be defined");

    const { data: user } = data;

    const roleMap: Record<EUserRole, string> = {
        [EUserRole.Admin]: "Admin",
        [EUserRole.Member]: "Member",
    };

    const links = [
        {
            route: ROUTE.userManagement.users.profile.$route,
            label: "General",
        },
        {
            route: "#",
            label: "API Keys",
        },
    ];

    return (
        <div className="bg-background pt-4 px-5 rounded-lg">
            <h3 className="text-2xl font-semibold">Your account</h3>

            <div className="flex items-center gap-4 mt-4 pb-3 border-b border-border">
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
                                <div className="flex gap-1">
                                    <span>Position :</span>
                                    <span className="text-black">{user.position || "Unknown position"}</span>
                                </div>
                            </div>
                            <span>•</span>
                            <div className="flex items-center gap-1.5">
                                <BadgeCheck className="size-4 text-blue-500" />
                                <div className="flex gap-1">
                                    <span>Role :</span>
                                    <span className="text-black">{roleMap[user.role] || user.role}</span>
                                </div>
                            </div>
                            {user.lastAccess && (
                                <>
                                    <span>•</span>
                                    <div className="flex items-center gap-1.5">
                                        <Clock className="size-4 text-blue-500" />
                                        <div className="flex gap-1">
                                            <span>Last access :</span>
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
            <TabNavigation links={links} />
        </div>
    );
}

interface Props {
    userId: string;
}
export const ProfileHeader = memo(View);
