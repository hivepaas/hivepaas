import { memo } from "react";

import { Avatar } from "@components/ui";
import { format } from "date-fns";
import { BadgeCheck, Clock, User } from "lucide-react";
import invariant from "tiny-invariant";

import { useProfileContext } from "@application/shared/context";

import { UserRoleBadge, UserStatusBadge } from "@application/modules/user-management/module-shared/components";

import { Separator } from "@/components/ui/separator";

function View() {
    const { profile } = useProfileContext();

    invariant(profile, "profile must be defined");

    return (
        <div className="bg-background py-3 sm:py-4 px-4 sm:px-5 rounded-lg">
            <h3 className="text-base sm:text-lg font-bold text-foreground pb-3">Your account</h3>

            <Separator className="opacity-50" />

            <div className="flex items-center gap-3 sm:gap-4 mt-3 sm:mt-4 pb-1">
                <Avatar
                    name={profile.fullName ?? profile.username}
                    src={profile.photo}
                    className="size-12 sm:size-16 md:size-20 text-lg sm:text-xl md:text-2xl shrink-0"
                />
                <div className="flex flex-col gap-1.5 sm:gap-2.5 min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                        <h2 className="text-base sm:text-lg font-semibold text-foreground truncate">
                            {profile.fullName ?? profile.username}
                        </h2>
                        <UserStatusBadge status={profile.status} />
                    </div>
                    <div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs sm:text-sm text-muted-foreground">
                        <div className="flex items-center gap-1.5">
                            <User className="size-3.5 sm:size-4 text-blue-500 shrink-0" />
                            <div className="flex gap-1 items-center">
                                <span>Position:</span>
                                <span className="text-foreground font-medium">
                                    {profile.position || "Unknown position"}
                                </span>
                            </div>
                        </div>
                        <span className="hidden sm:inline">•</span>
                        <div className="flex gap-1.5 items-center">
                            <BadgeCheck className="size-3.5 sm:size-4 text-blue-500 shrink-0" />
                            <div className="flex gap-1 items-center">
                                <span>Role:</span>
                                <UserRoleBadge role={profile.role} />
                            </div>
                        </div>
                        {profile.lastAccess && (
                            <>
                                <span className="hidden sm:inline">•</span>
                                <div className="flex items-center gap-1.5">
                                    <Clock className="size-3.5 sm:size-4 text-blue-500 shrink-0" />
                                    <div className="flex gap-1">
                                        <span>Last access:</span>
                                        <span className="text-foreground font-medium">
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
    );
}

export const ProfileHeader = memo(View);
