import { useInstallationTimezone } from "@application/shared/hooks";

import { FieldDescription } from "@/components/ui";

/**
 * Says what a cron expression's hours are read in: the installation's timezone.
 */
export function CronTimezoneNote() {
    const timezone = useInstallationTimezone();

    return <FieldDescription>Its hours are in {timezone}, the installation&apos;s timezone.</FieldDescription>;
}
