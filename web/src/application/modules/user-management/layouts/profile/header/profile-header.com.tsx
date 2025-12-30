import { memo } from "react";

import { Avatar } from "@components/ui";
import { moduleHeaderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { format } from "date-fns";
import { BadgeCheck, Clock, User } from "lucide-react";
import invariant from "tiny-invariant";

import { TabNavigation } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";

import { UserRoleBadge, UserStatusBadge } from "@application/modules/user-management/module-shared/components";

function View() {
    const { profile } = useProfileContext();

    invariant(profile, "profile must be defined");

    const links = [
        {
            route: ROUTE.currentUser.profile.$route,
            label: "General",
        },
        {
            route: ROUTE.currentUser.profileApiKeys.$route,
            label: "API Keys",
        },
    ];

    return (
        <div className={cn(moduleHeaderBox)}>
            <h3 className="text-lg font-semibold">Your account</h3>

            <div className="flex items-center gap-4 mt-4 pb-3 border-b border-border">
                <div className="flex items-center gap-4">
                    <Avatar
                        name={profile.fullName ?? ""}
                        src={profile.photo}
                        className="size-20 text-2xl"
                    />
                    <div className="flex flex-col gap-3">
                        <div className="flex items-center gap-2">
                            <h2 className="text-lg font-semibold text-foreground">{profile.fullName}</h2>
                            <UserStatusBadge status={profile.status} />
                        </div>
                        <div className="flex items-center gap-4 text-sm text-muted-foreground">
                            <div className="flex items-center gap-1.5">
                                <User className="size-4 text-blue-500" />
                                <div className="flex gap-1">
                                    <span>Position:</span>
                                    <span className="text-black">{profile.position || "Unknown position"}</span>
                                </div>
                            </div>
                            <span>•</span>
                            <div className="flex gap-1.5 items-center">
                                <BadgeCheck className="size-4 text-blue-500" />
                                <div className="flex gap-1">
                                    <span>Role:</span>
                                    <UserRoleBadge role={profile.role} />
                                </div>
                            </div>
                            {profile.lastAccess && (
                                <>
                                    <span>•</span>
                                    <div className="flex items-center gap-1.5">
                                        <Clock className="size-4 text-blue-500" />
                                        <div className="flex gap-1">
                                            <span>Last access:</span>
                                            <span className="text-black">
                                                {format(profile.lastAccess, "yyyy-MM-dd HH:mm:ss")}{" "}
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

export const ProfileHeader = memo(View);
