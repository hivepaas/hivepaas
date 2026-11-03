import { FieldDescription } from "@/components/ui";

/**
 * Says what a date and time picked here is read in: this browser's timezone, which may not be the installation's
 * that next runs and cron hours are shown in.
 */
export function BrowserTimezoneNote() {
    const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

    return <FieldDescription>In {timezone}, this browser&apos;s timezone.</FieldDescription>;
}
