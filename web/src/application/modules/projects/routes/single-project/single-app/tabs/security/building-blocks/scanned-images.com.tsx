import { AlertTriangle, Box } from "lucide-react";
import type { ScannedImage } from "~/projects/api/services";

import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Badge } from "@/components/ui/badge";

import { shortImageId } from "./severity";

/** The images the app's tasks ran when last scanned, each with its last scan. */
export function ScannedImages({ images }: { images: ScannedImage[] }) {
    return (
        <div className="flex flex-col gap-2">
            {images.map(image => (
                <div
                    key={image.imageId}
                    data-testid="scanned-image"
                    className="flex flex-col gap-1.5 rounded-lg border px-3 py-2.5 text-[13px]"
                >
                    <div className="flex flex-wrap items-center gap-2">
                        <Box className="size-4 shrink-0 text-muted-foreground" />
                        <span className="font-mono font-medium break-all">{image.imageRef || "(no tag)"}</span>
                        <span
                            className="font-mono text-xs text-muted-foreground"
                            title={image.imageId}
                        >
                            {shortImageId(image.imageId)}
                        </span>
                        {image.platform && <Badge tone="neutral">{image.platform}</Badge>}
                        {image.osName && (
                            <Badge tone="indigo">
                                {image.osFamily} {image.osName}
                            </Badge>
                        )}
                        {image.status === "failed" && (
                            <Badge
                                tone="red"
                                appearance="solid"
                            >
                                Scan failed
                            </Badge>
                        )}
                    </div>
                    {image.osEosl && (
                        <div className="flex items-start gap-1.5 text-amber-700 dark:text-amber-400">
                            <AlertTriangle className="mt-0.5 size-3.5 shrink-0" />
                            <span>
                                {image.osFamily} {image.osName} is past its support: its packages get no more fixes, and
                                what the scan cannot find may still be there. Move the image to a supported release.
                            </span>
                        </div>
                    )}
                    {image.error && <div className="text-destructive break-words">{image.error}</div>}
                    <div className="text-xs text-muted-foreground">
                        Scanned {timeAgoFormatter.format(image.scannedAt)}
                        {image.source === "registry" && <>, read from its registry</>}
                        {image.dbUpdatedAt && <> · vulnerability DB of {timeAgoFormatter.format(image.dbUpdatedAt)}</>}
                        {image.trivyVersion && <> · Trivy {image.trivyVersion}</>}
                    </div>
                </div>
            ))}
        </div>
    );
}
