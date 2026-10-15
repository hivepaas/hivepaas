import type { ImageScanVolume, ImageScannerStatus } from "~/system-settings/domain";
import { SectionHeader } from "~/system-settings/module-shared";

import { InfoBlock } from "@application/shared/components";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

/** Where scans run, by the volume's pin. */
function whereScansRun(volume: ImageScanVolume): string {
    if (volume.nodeId) {
        return `On node ${volume.nodeName || volume.nodeId}, where the volume is`;
    }
    if (volume.nodeLabel) {
        return `On a node carrying the label ${volume.nodeLabel}`;
    }
    return "On the node running each image: the volume is shared";
}

/**
 * Where scans run, as the chosen volume decides, and what the last one used:
 * Trivy's image and the vulnerability DB it was matched against.
 */
export function ImageScannerStatusSection({ status, enabled }: { status: ImageScannerStatus; enabled: boolean }) {
    if (!enabled && !status.lastScanAt) {
        return null;
    }
    const { volume } = status;

    return (
        <div
            className="flex flex-col gap-4"
            data-testid="image-scanner-status"
        >
            <SectionHeader>Status</SectionHeader>
            <div className="flex flex-col gap-4 px-3 text-sm">
                {volume && (
                    <>
                        <InfoBlock
                            titleWidth={220}
                            title="Data volume"
                        >
                            {volume.missing ? (
                                <span className="text-amber-700 dark:text-amber-400">
                                    The volume is gone: no scan can run. Choose another below.
                                </span>
                            ) : (
                                <span>{volume.name}</span>
                            )}
                        </InfoBlock>
                        {!volume.missing && (
                            <InfoBlock
                                titleWidth={220}
                                title="Scans run"
                            >
                                <span className="flex flex-col gap-1">
                                    <span>{whereScansRun(volume)}</span>
                                    {volume.problem && (
                                        <span
                                            className="text-amber-700 dark:text-amber-400"
                                            data-testid="image-scan-volume-problem"
                                        >
                                            {volume.problem}
                                        </span>
                                    )}
                                </span>
                            </InfoBlock>
                        )}
                    </>
                )}
                <InfoBlock
                    titleWidth={220}
                    title="Image"
                >
                    <span className="font-mono text-xs break-all">{status.image}</span>
                </InfoBlock>
                {status.lastScanAt && (
                    <InfoBlock
                        titleWidth={220}
                        title="Last scan"
                    >
                        <span>
                            {timeAgoFormatter.format(status.lastScanAt)}
                            {status.trivyVersion && (
                                <span className="text-muted-foreground"> · Trivy {status.trivyVersion}</span>
                            )}
                        </span>
                    </InfoBlock>
                )}
                {status.lastError && (
                    <InfoBlock
                        titleWidth={220}
                        title="Last error"
                    >
                        <span className="text-amber-700 dark:text-amber-400 whitespace-pre-wrap break-words">
                            {status.lastError}
                        </span>
                    </InfoBlock>
                )}
                {status.dbUpdatedAt && (
                    <InfoBlock
                        titleWidth={220}
                        title="Vulnerability DB"
                    >
                        <span>Built {timeAgoFormatter.format(status.dbUpdatedAt)}</span>
                    </InfoBlock>
                )}
            </div>
        </div>
    );
}
