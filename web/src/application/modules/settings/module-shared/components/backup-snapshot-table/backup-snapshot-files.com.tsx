import { useState } from "react";

import { ChevronRight, Download, File, Folder, Loader2 } from "lucide-react";
import { useBackupSnapshotApi } from "~/settings/api/hooks";
import { BackupSnapshotQueries } from "~/settings/data/queries";
import type { BackupSnapshotScope } from "~/settings/domain";

import { formatDataSizeCompact } from "@application/shared/utils/data-size";

import { Button } from "@/components/ui";

import { joinSnapshotPath } from "./backup-snapshot-restore.helpers";

function saveBlob(blob: Blob, filename: string) {
    const url = window.URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    anchor.click();
    window.URL.revokeObjectURL(url);
}

/**
 * What a snapshot holds, browsed a directory at a time, each file with its download. A download is held whole
 * in the browser's memory before it is saved.
 */
export function BackupSnapshotFiles({ scope, snapshotRecordId, canDownload }: Props) {
    const [browsed, setBrowsed] = useState("");
    const [downloading, setDownloading] = useState("");
    const { queries } = useBackupSnapshotApi();
    const { data, isFetching } = BackupSnapshotQueries.useFindEntries({ scope, id: snapshotRecordId, path: browsed });
    const entries = data?.data ?? [];
    const crumbs = browsed ? browsed.split("/") : [];

    async function download(path: string) {
        setDownloading(path);
        try {
            const { data: file } = await queries.downloadFile({ scope, id: snapshotRecordId, path });
            saveBlob(file.blob, file.filename);
        } catch {
            // The error is notified by the query.
        } finally {
            setDownloading("");
        }
    }

    return (
        <div className="rounded-md border border-border/80">
            <div className="flex flex-wrap items-center gap-1 border-b border-border/60 px-2 py-1.5 text-xs">
                <Button
                    type="button"
                    variant="link"
                    size="sm"
                    className="h-auto p-0 text-xs"
                    onClick={() => {
                        setBrowsed("");
                    }}
                >
                    /
                </Button>
                {crumbs.map((crumb, index) => {
                    const path = crumbs.slice(0, index + 1).join("/");
                    return (
                        <span
                            key={path}
                            className="flex items-center gap-1"
                        >
                            <ChevronRight className="size-3 text-muted-foreground" />
                            <Button
                                type="button"
                                variant="link"
                                size="sm"
                                className="h-auto p-0 font-mono text-xs"
                                onClick={() => {
                                    setBrowsed(path);
                                }}
                            >
                                {crumb}
                            </Button>
                        </span>
                    );
                })}
                {isFetching && <Loader2 className="ml-auto size-3 animate-spin text-muted-foreground" />}
            </div>
            <ul className="max-h-72 overflow-y-auto py-1 text-sm">
                {!isFetching && entries.length === 0 && <li className="px-3 py-1.5 text-muted-foreground">Empty</li>}
                {entries.map(entry => {
                    const path = joinSnapshotPath(browsed, entry.name);
                    return (
                        <li
                            key={path}
                            className="flex items-center gap-2 px-3 py-1"
                        >
                            {entry.dir ? (
                                <Folder className="size-4 shrink-0 text-muted-foreground" />
                            ) : (
                                <File className="size-4 shrink-0 text-muted-foreground" />
                            )}
                            {entry.dir ? (
                                <button
                                    type="button"
                                    className="truncate font-mono text-left hover:underline cursor-pointer"
                                    onClick={() => {
                                        setBrowsed(path);
                                    }}
                                >
                                    {entry.name}/
                                </button>
                            ) : (
                                <span className="truncate font-mono">{entry.name}</span>
                            )}
                            <span className="ml-auto shrink-0 text-xs text-muted-foreground">
                                {formatDataSizeCompact(entry.sizeBytes)}
                            </span>
                            {!entry.dir && canDownload && (
                                <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="size-7 shrink-0"
                                    aria-label={`Download ${path}`}
                                    disabled={downloading !== ""}
                                    onClick={() => {
                                        void download(path);
                                    }}
                                >
                                    {downloading === path ? (
                                        <Loader2 className="size-4 animate-spin" />
                                    ) : (
                                        <Download className="size-4" />
                                    )}
                                </Button>
                            )}
                        </li>
                    );
                })}
            </ul>
        </div>
    );
}

interface Props {
    scope: BackupSnapshotScope;
    snapshotRecordId: string;
    /** Whether the viewer may take a file: write on the snapshot's owner. */
    canDownload: boolean;
}
