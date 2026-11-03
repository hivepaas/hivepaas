import { AlertTriangle, CircleCheck, CircleX } from "lucide-react";
import { Link } from "react-router";
import { type AttentionItem, AttentionScope } from "~/home/domain";

import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

import { describeAttentionItem } from "./attention-item-text";

interface Props {
    items: AttentionItem[];
    isLoading: boolean;
}

function AttentionRow({ item }: { item: AttentionItem }) {
    const { title, detail, link } = describeAttentionItem(item);
    const critical = item.severity === "critical";
    const where = item.scope === AttentionScope.App && item.project ? `${item.project.name} / ${item.env}` : undefined;

    return (
        <li className="flex items-start gap-3 px-5 py-3.5 border-b border-border/60 last:border-b-0">
            {critical ? (
                <CircleX
                    className="mt-0.5 size-4 shrink-0 text-destructive"
                    aria-label="Critical"
                />
            ) : (
                <AlertTriangle
                    className="mt-0.5 size-4 shrink-0 text-amber-600 dark:text-amber-400"
                    aria-label="Warning"
                />
            )}
            <div className="flex min-w-0 grow flex-col gap-1">
                <span className="text-sm font-semibold">
                    {title}
                    {where && <span className="font-normal text-muted-foreground"> · {where}</span>}
                </span>
                <span className="text-[13px] text-muted-foreground break-words">{detail}</span>
            </div>
            <div className="flex shrink-0 flex-col items-end gap-1">
                {link && (
                    <Link
                        to={link.to}
                        className="text-[13px] font-medium text-foreground underline-offset-4 hover:underline"
                    >
                        {link.label}
                    </Link>
                )}
                {item.since && (
                    <span className="text-xs text-muted-foreground">{timeAgoFormatter.format(item.since)}</span>
                )}
            </div>
        </li>
    );
}

/**
 * What needs someone, across everything the user may see. The server has
 * already left out what they may not, and says nothing about it.
 */
export function NeedsAttentionCard({ items, isLoading }: Props) {
    return (
        <Card className="gap-0 py-0">
            <CardHeader className="flex flex-row items-center gap-2 border-b px-5 py-4 [.border-b]:pb-4">
                <CardTitle className="text-[15px]">Needs attention</CardTitle>
                {items.length > 0 && (
                    <Badge
                        variant="destructive"
                        className="rounded-full px-2"
                    >
                        {items.length}
                    </Badge>
                )}
            </CardHeader>
            <CardContent className="px-0">
                {isLoading ? (
                    <div className="flex flex-col gap-3 p-5">
                        <Skeleton className="h-10 w-full" />
                        <Skeleton className="h-10 w-full" />
                    </div>
                ) : items.length === 0 ? (
                    <div className="flex items-center gap-3 px-5 py-6 text-sm text-muted-foreground">
                        <CircleCheck className="size-5 text-green-600 dark:text-green-500" />
                        Nothing needs attention right now.
                    </div>
                ) : (
                    <ul>
                        {items.map(item => (
                            <AttentionRow
                                key={`${item.kind}:${item.scope}:${item.app?.id ?? item.subject}`}
                                item={item}
                            />
                        ))}
                    </ul>
                )}
            </CardContent>
        </Card>
    );
}
