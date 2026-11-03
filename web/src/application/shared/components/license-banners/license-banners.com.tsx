import { ROUTE } from "@application/shared/constants";
import { useReadableLicense } from "@application/shared/hooks";
import { useAppMatchPath } from "@application/shared/hooks/router";

import { isLicenseBannerDismissed, useLicenseBannerDismissalsStore } from "./license-banner-dismissals.store";
import { LicenseBanner } from "./license-banner.com";
import { licenseBannerSnoozeText, licenseBannersOf } from "./license-banners.utils";

/**
 * Tells whoever can read System - every admin, and any member given System -
 * what the license means for this installation: a trial or a grace period
 * running out, a key that no longer puts its plan in force, or more of a kind
 * than the plan allows. See licenseBannersOf for which, and what they say.
 *
 * Anyone else is not told, and the license is not even asked for: the server
 * would refuse them. Nor is anything said when it cannot be read - the read is
 * silent, as a toast following someone from page to page about something they
 * did not ask for would be noise; System › License shows the error.
 *
 * System › License says what the status banners say already, so there only the
 * limits banner is shown, without its link to the page it is on.
 */
export function LicenseBanners() {
    // Once at sign-in and every half hour while the dashboard is open, besides
    // whenever the license page, the plan-limit dialog or a change to the key
    // reads it again: they all share the one query.
    const { license } = useReadableLicense({ staleTime: 5 * 60_000, refetchInterval: 30 * 60_000 });
    const { match } = useAppMatchPath();
    const dismissed = useLicenseBannerDismissalsStore(state => state.dismissed);
    const dismiss = useLicenseBannerDismissalsStore(state => state.dismiss);

    if (!license) {
        return null;
    }

    // Read at each render, which every page change brings: a dismissal that has
    // run out shows its banner again on the next page.
    const now = Date.now();
    const onLicensePage = match.modules({ pattern: ROUTE.systemSettings.license.$pattern }) !== null;
    const banners = licenseBannersOf(license, now).filter(
        banner => !(onLicensePage && banner.kind === "status") && !isLicenseBannerDismissed(dismissed, banner.id, now),
    );

    if (banners.length === 0) {
        return null;
    }

    return (
        <div className="flex flex-col gap-2">
            {banners.map(banner => (
                <LicenseBanner
                    key={banner.id}
                    tone={banner.tone}
                    title={banner.title}
                    body={banner.body}
                    links={onLicensePage ? banner.links.filter(link => link.external) : banner.links}
                    dismissHint={`Hide for ${licenseBannerSnoozeText(banner.snoozeMs)}`}
                    onDismiss={() => {
                        dismiss(banner.id, banner.snoozeMs);
                    }}
                />
            ))}
        </div>
    );
}
