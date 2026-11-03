import { CheckCircle2Icon, InfoIcon } from "lucide-react";
import type { SpecExportResult } from "~/operations/domain";

/**
 * Issue codes grouped by what the reader would do about them, rather than
 * listed raw. The codes themselves are the server's vocabulary, not a user's.
 */
const CODE_GROUPS: { title: string; codes: string[]; note: string }[] = [
    {
        title: "Left out on purpose",
        codes: ["TYPE_SKIPPED", "PREVIEW_APP_SKIPPED"],
        note: "Settings that cannot be re-created from a file, and preview environments, which belong to the pull request that made them.",
    },
    {
        title: "Nothing to record yet",
        codes: ["SERVICE_UNAVAILABLE"],
        note: "Apps that have never been deployed. Their settings are in the bundle; how they run is not, because there is nothing running.",
    },
    {
        title: "Needs your attention",
        codes: ["REF_NOT_FOUND", "REF_NOT_SELECTED", "TYPE_UNCLASSIFIED", "MOUNT_TARGET_DUPLICATED"],
        note: "References this scope could not resolve. Importing elsewhere will ask you for them.",
    },
];

function formatSize(bytes: number): string {
    if (bytes < 1024) {
        return `${bytes} B`;
    }
    if (bytes < 1024 * 1024) {
        return `${(bytes / 1024).toFixed(1)} KB`;
    }

    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function SpecExportResultPanel({ result }: Props) {
    const { summary } = result;
    const byCode = summary?.byCode ?? {};

    const groups = CODE_GROUPS.map(group => ({
        ...group,
        count: group.codes.reduce((total, code) => total + (byCode[code] ?? 0), 0),
    })).filter(group => group.count > 0);

    return (
        <div className="rounded-lg border bg-background p-4">
            <div className="flex flex-col gap-4">
                <div className="flex items-start gap-2">
                    <CheckCircle2Icon className="mt-0.5 size-5 shrink-0 text-emerald-600" />
                    <div>
                        <p className="text-sm font-medium text-foreground">Saved {result.filename}</p>
                        <p className="text-xs text-muted-foreground">
                            {formatSize(result.sizeBytes)}
                            {summary ? ` · ${summary.files} file${summary.files === 1 ? "" : "s"}` : ""}
                        </p>
                    </div>
                </div>

                {groups.length > 0 && (
                    <div className="flex flex-col gap-3 border-t pt-4">
                        {groups.map(group => (
                            <div
                                key={group.title}
                                className="flex items-start gap-2"
                            >
                                <InfoIcon className="mt-0.5 size-4 shrink-0 text-muted-foreground" />
                                <div>
                                    <p className="text-sm text-foreground">
                                        {group.title}: <span className="font-medium">{group.count}</span>
                                    </p>
                                    <p className="text-xs text-muted-foreground">{group.note}</p>
                                </div>
                            </div>
                        ))}
                        {summary?.reportFile && (
                            <p className="text-xs text-muted-foreground">
                                The full list is in <code className="font-mono">{summary.reportFile}</code> inside the
                                archive.
                            </p>
                        )}
                    </div>
                )}
            </div>
        </div>
    );
}

interface Props {
    result: SpecExportResult;
}
