import { useState } from "react";

import { HivePaaSUpdatesQueries } from "~/system-settings/data";
import type { ReleaseChannel } from "~/system-settings/domain";

import { MODULE_IDS } from "@application/shared/constants";
import { useProfileContext } from "@application/shared/context";
import { EUserRole } from "@application/shared/enums";
import { useConditionalModule } from "@application/shared/permissions";

import { Skeleton } from "@/components/ui/skeleton";

import {
    ReleaseChannelCard,
    ReviewUpdateDialog,
    RunningReleaseCard,
    UpdateHistoryCard,
    UpdatingOverlay,
} from "../building-blocks";

/**
 * The release running, the latest on each channel, and the way to move to one.
 * Updating is an admin's to do, as the server requires.
 */
export function SystemSettingsHivePaaSUpdatesRoute() {
    const profile = useProfileContext(state => state.profile);
    const isAdmin = profile?.role === EUserRole.Admin;
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });

    const { data, isLoading, isFetching, dataUpdatedAt, refetch } = HivePaaSUpdatesQueries.useFindReleaseInfo({
        enabled: isAdmin,
    });
    const [reviewing, setReviewing] = useState<ReleaseChannel | null>(null);
    const [updatingTo, setUpdatingTo] = useState<string | null>(null);

    if (!isAdmin) {
        return <p className="text-sm text-muted-foreground">Only an admin can see and apply HivePaaS updates.</p>;
    }

    const info = data?.data;

    return (
        <div className="flex flex-col gap-5">
            {isLoading || !info ? (
                <Skeleton className="h-24 w-full" />
            ) : (
                <>
                    <RunningReleaseCard
                        current={info.current}
                        checkedAt={dataUpdatedAt}
                        isChecking={isFetching}
                        onCheck={() => void refetch()}
                    />
                    <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
                        {(["stable", "beta"] as const).map(channel => (
                            <ReleaseChannelCard
                                key={channel}
                                channel={channel}
                                release={info[channel]}
                                canManage={canWrite}
                                onReview={() => {
                                    setReviewing(channel);
                                }}
                            />
                        ))}
                    </div>
                    <ReviewUpdateDialog
                        open={reviewing !== null}
                        channel={reviewing ?? "stable"}
                        releaseInfo={info}
                        onOpenChange={open => {
                            if (!open) {
                                setReviewing(null);
                            }
                        }}
                        onStarted={version => {
                            setReviewing(null);
                            setUpdatingTo(version);
                        }}
                    />
                </>
            )}

            <UpdateHistoryCard />

            {updatingTo && <UpdatingOverlay version={updatingTo} />}
        </div>
    );
}
