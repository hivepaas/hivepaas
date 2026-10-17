import { AppLoader } from "@application/shared/components";
import { MODULE_IDS } from "@application/shared/constants";
import { LicenseQueries } from "@application/shared/data/queries";
import { PageError } from "@application/shared/pages";
import { useConditionalModule } from "@application/shared/permissions";

import { InstallationIdSection, LicensePlanCard, LicenseRefreshSection, LicenseUsageSection } from "../building-blocks";
import { LicenseKeyForm } from "../form";

/**
 * The license in force on this installation: its plan and status, how much of
 * each limit is used, the installation ID a key is bought for, and where a key
 * is pasted or a newer one asked for.
 */
export function SystemSettingsLicenseRoute() {
    const { canWrite } = useConditionalModule({ id: MODULE_IDS.System });
    const { data, isLoading, error, refetch } = LicenseQueries.useFindOne();
    const license = data?.data;

    if (isLoading) {
        return <AppLoader />;
    }

    // The query never toasts - it also runs on every page for the banners - so a
    // failure is said here. A later refetch that fails keeps what was shown.
    if (!license) {
        return (
            <PageError
                error={error ?? new Error("The license could not be read.")}
                onRetry={refetch}
            />
        );
    }

    return (
        <div className="flex flex-col gap-6 pt-2">
            <LicensePlanCard license={license} />
            <LicenseUsageSection license={license} />
            <InstallationIdSection installationId={license.installationId} />
            <LicenseKeyForm
                license={license}
                canWrite={canWrite}
            />
            <LicenseRefreshSection license={license} />
        </div>
    );
}
