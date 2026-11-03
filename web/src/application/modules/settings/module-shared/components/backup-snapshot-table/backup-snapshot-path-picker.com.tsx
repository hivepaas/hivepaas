import { useState } from "react";

import { cn } from "@lib/utils";
import { ChevronRight, File, Folder, Loader2 } from "lucide-react";
import { BackupSnapshotQueries } from "~/settings/data/queries";
import type { BackupSnapshotScope } from "~/settings/domain";

import { formatDataSizeCompact } from "@application/shared/utils/data-size";

import { Button } from "@/components/ui";

import { joinSnapshotPath, parentSnapshotPath } from "./backup-snapshot-restore.helpers";

/**
 * Picks what of a snapshot to restore: all of it, or one of its directories, browsed a directory at a time.
 * `onChange` is told the path and its size.
 */
export function BackupSnapshotPathPicker({ scope, snapshotRecordId, snapshotSize, value, onChange, disabled }: Props) {
    const [browsed, setBrowsed] = useState(() => parentSnapshotPath(value));
    const { data, isFetching } = BackupSnapshotQueries.useFindEntries({
        scope,
        id: snapshotRecordId,
        path: browsed,
    });
    const entries = data?.data ?? [];
    const crumbs = browsed ? browsed.split("/") : [];

    return (
        <div className="flex flex-col gap-2">
            <label className="flex items-center gap-2 text-sm cursor-pointer">
                <input
                    type="radio"
                    checked={value === ""}
                    disabled={disabled}
                    onChange={() => {
                        onChange("", snapshotSize);
                    }}
                />
                <span>All of it</span>
                <span className="text-muted-foreground">({formatDataSizeCompact(snapshotSize)})</span>
            </label>

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
                <ul className="max-h-56 overflow-y-auto py-1 text-sm">
                    {!isFetching && entries.length === 0 && (
                        <li className="px-3 py-1.5 text-muted-foreground">Empty</li>
                    )}
                    {entries.map(entry => {
                        const path = joinSnapshotPath(browsed, entry.name);
                        return (
                            <li
                                key={path}
                                className={cn(
                                    "flex items-center gap-2 px-3 py-1",
                                    !entry.dir && "text-muted-foreground",
                                )}
                            >
                                {entry.dir ? (
                                    <input
                                        type="radio"
                                        aria-label={`Restore ${path}`}
                                        checked={value === path}
                                        disabled={disabled}
                                        onChange={() => {
                                            onChange(path, entry.sizeBytes);
                                        }}
                                    />
                                ) : (
                                    <span className="w-[13px]" />
                                )}
                                {entry.dir ? (
                                    <Folder className="size-4 shrink-0 text-muted-foreground" />
                                ) : (
                                    <File className="size-4 shrink-0" />
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
                            </li>
                        );
                    })}
                </ul>
            </div>
        </div>
    );
}

interface Props {
    scope: BackupSnapshotScope;
    snapshotRecordId: string;
    snapshotSize: number;
    /** A directory inside the snapshot; "" for all of it. */
    value: string;
    onChange: (path: string, sizeBytes: number) => void;
    disabled?: boolean;
}
