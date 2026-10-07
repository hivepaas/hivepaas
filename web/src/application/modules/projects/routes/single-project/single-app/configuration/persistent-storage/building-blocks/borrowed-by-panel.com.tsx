import { TriangleAlert } from "lucide-react";
import type { MountBorrower } from "~/projects/domain";

import { ContentBlock } from "@application/shared/components";

interface Props {
    borrowers: MountBorrower[];
}

// The API omits an empty subpath, so a missing one means the whole directory.

/**
 * Who else reaches this app's files.
 *
 * The other side of the arrangement is shown on the borrowing app's own screen,
 * where the person who set it up is looking. This is for the person who owns the
 * data, and who would otherwise have no way of learning that anything reads it -
 * a grant only the borrower can see is a grant nobody will ever revoke.
 */
export function BorrowedByPanel({ borrowers }: Props) {
    if (borrowers.length === 0) {
        return null;
    }

    return (
        <ContentBlock label="Apps reading this app's storage">
            <div className="flex flex-col gap-2">
                {borrowers.map(borrower => (
                    <div
                        key={`${borrower.appId}:${borrower.target}`}
                        className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm"
                    >
                        <TriangleAlert
                            className={
                                borrower.write
                                    ? "size-4 shrink-0 text-amber-600 dark:text-amber-400"
                                    : "size-4 shrink-0 text-muted-foreground"
                            }
                        />
                        <span className="font-medium">{borrower.name}</span>
                        <span className="text-muted-foreground">{borrower.write ? "reads and changes" : "reads"}</span>
                        <code className="rounded bg-muted px-1 py-0.5 text-xs">{borrower.subpath ?? "everything"}</code>
                        <span className="text-xs text-muted-foreground">at {borrower.target} in that app</span>
                    </div>
                ))}
                <p className="text-xs text-muted-foreground">
                    Remove the mount on that app&apos;s Persistent Storage screen to take this away.
                </p>
            </div>
        </ContentBlock>
    );
}
