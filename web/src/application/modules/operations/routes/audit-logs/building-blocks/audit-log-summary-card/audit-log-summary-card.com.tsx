import { type KeyboardEvent, useEffect } from "react";

import { cn } from "@/lib/utils";
import { format } from "date-fns";
import { Calendar, ChevronDown, FolderTree, Globe, KeyRound, Layers, Loader2 } from "lucide-react";
import ReactTimeAgo from "react-time-ago";
import { AuditLogsQueries } from "~/operations/data";
import { type AuditLog, AuditLogResult, type AuditLogScope } from "~/operations/domain";

import { Avatar } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";

const RESULT_BORDER_CLASS_NAMES: Record<string, string> = {
    [AuditLogResult.Allowed]: "border-l-green-500",
    [AuditLogResult.Denied]: "border-l-red-500",
};

const RESULT_BADGE_CLASS_NAMES: Record<string, string> = {
    [AuditLogResult.Allowed]: "bg-green-500 text-white hover:bg-green-500/90",
    [AuditLogResult.Denied]: "bg-red-500 text-white hover:bg-red-500/90",
};

function formatDateTime(date: Date | null): string {
    return date ? format(date, "yyyy-MM-dd HH:mm:ss") : "-";
}

function formatDetail(detail: string): string {
    try {
        const parsed: unknown = JSON.parse(detail);
        return JSON.stringify(parsed, null, 2);
    } catch {
        return detail;
    }
}

export function ResultBadge({ result }: { result: string }) {
    const isAllowed = result.toLowerCase() === AuditLogResult.Allowed;
    const isDenied = result.toLowerCase() === AuditLogResult.Denied;

    const className = isAllowed
        ? RESULT_BADGE_CLASS_NAMES[AuditLogResult.Allowed]
        : isDenied
          ? RESULT_BADGE_CLASS_NAMES[AuditLogResult.Denied]
          : "bg-slate-500 text-white hover:bg-slate-500/90";

    const label = isAllowed ? "Allowed" : isDenied ? "Denied" : result;

    return (
        <Badge
            variant="default"
            className={cn("px-2.5 py-0.5 text-xs font-semibold capitalize tracking-wide shadow-2xs", className)}
        >
            {label}
        </Badge>
    );
}

export interface AuditLogSummaryCardProps {
    log: AuditLog;
    scope?: AuditLogScope;
    isExpanded?: boolean;
    cachedDetail?: AuditLog;
    onDetailLoaded?: (detail: AuditLog) => void;
    onToggleExpand?: () => void;
    onClick?: () => void;
}

export function AuditLogSummaryCard({
    log,
    scope,
    isExpanded = false,
    cachedDetail,
    onDetailLoaded,
    onToggleExpand,
    onClick,
}: AuditLogSummaryCardProps) {
    const handleClick = onToggleExpand ?? onClick;
    const isClickable = Boolean(handleClick);

    const hasDetailAlready = cachedDetail !== undefined || log.detail !== undefined;
    const shouldFetchDetail = isExpanded && Boolean(log.hasDetail) && !hasDetailAlready;

    const { data: detailResponse, isFetching: isFetchingDetail } = AuditLogsQueries.useFindOneById(
        { scope, itemID: log.id },
        { enabled: shouldFetchDetail },
    );

    useEffect(() => {
        if (detailResponse?.data && onDetailLoaded) {
            onDetailLoaded(detailResponse.data);
        }
    }, [detailResponse?.data, onDetailLoaded]);

    const detailedLog = cachedDetail ?? detailResponse?.data ?? log;

    const borderClass = RESULT_BORDER_CLASS_NAMES[detailedLog.result.toLowerCase()] ?? "border-l-slate-400";

    const actorName = detailedLog.actor?.name ?? detailedLog.actor?.loggedName ?? detailedLog.actor?.id ?? "System";
    const ipAddress = detailedLog.clientIp ?? detailedLog.remoteAddr ?? "";
    const resourceName = detailedLog.resource?.name ?? detailedLog.resource?.loggedName ?? detailedLog.resource?.id;

    const isProjectOrEnvScope = scope?.type === "project" || scope?.type === "project-env";
    const showSummaryScope = isProjectOrEnvScope
        ? Boolean(detailedLog.scopeApp)
        : Boolean(detailedLog.scopeProject ?? detailedLog.scopeApp ?? detailedLog.scopeUser);

    const hasAnyScope = Boolean(detailedLog.scopeProject ?? detailedLog.scopeApp ?? detailedLog.scopeUser);

    function handleKeyDown(event: KeyboardEvent<HTMLElement>) {
        if (!handleClick || (event.key !== "Enter" && event.key !== " ")) {
            return;
        }

        event.preventDefault();
        handleClick();
    }

    return (
        <div
            className={cn(
                "rounded-[8px] border bg-background shadow-xs p-3 sm:p-4 transition-all duration-200",
                borderClass,
                isExpanded && "ring-1 ring-border shadow-sm",
            )}
            style={{
                borderLeftWidth: 4,
            }}
        >
            {/* Clickable Card Summary: Rows 1-3 */}
            <div
                className={cn(
                    "flex flex-col gap-2.5",
                    isClickable &&
                        "cursor-pointer hover:opacity-95 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded-sm",
                )}
                role={isClickable ? "button" : undefined}
                tabIndex={isClickable ? 0 : undefined}
                onClick={handleClick}
                onKeyDown={handleKeyDown}
            >
                {/* Line 1: Header - Result badge, Type badge, Source badge, Timestamp & Expand Chevron */}
                <div className="flex flex-wrap items-center justify-between gap-3">
                    <div className="flex flex-wrap items-center gap-2 sm:gap-2.5">
                        <ResultBadge result={detailedLog.result} />

                        <Badge
                            variant="outline"
                            className="font-mono text-xs px-2 py-0.5 rounded-md border-border/70 bg-muted/50 text-foreground"
                        >
                            {detailedLog.type}
                        </Badge>

                        {detailedLog.source && (
                            <Badge
                                variant="secondary"
                                className="font-mono text-xs px-2 py-0.5 rounded-md text-muted-foreground bg-muted/40 border border-border/40"
                            >
                                {detailedLog.source}
                            </Badge>
                        )}
                    </div>

                    <div className="flex items-center gap-3">
                        <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                            <Calendar className="size-4 shrink-0 text-muted-foreground/70" />
                            <span className="font-medium text-foreground">{formatDateTime(detailedLog.createdAt)}</span>
                            <span className="text-muted-foreground/60 text-xs">
                                (
                                <ReactTimeAgo
                                    date={detailedLog.createdAt}
                                    locale="en-US"
                                />
                                )
                            </span>
                        </div>

                        {isClickable && (
                            <ChevronDown
                                className={cn(
                                    "size-4 text-muted-foreground transition-transform duration-200 shrink-0",
                                    isExpanded && "rotate-180 text-foreground",
                                )}
                            />
                        )}
                    </div>
                </div>

                {/* Line 2: Actor photo + name + viaApiKey + IP / remoteAddr */}
                <div className="flex flex-wrap items-center gap-x-3 gap-y-1.5 text-sm text-muted-foreground pt-2 border-t border-border/50">
                    <div className="flex items-center gap-2">
                        <Avatar
                            name={actorName}
                            src={detailedLog.actor?.photo}
                            className="size-5 text-[10px] border border-border"
                        />
                        <span className="font-medium text-foreground">{actorName}</span>
                        {detailedLog.actor?.type && detailedLog.actor.type !== "user" && (
                            <span className="text-xs text-muted-foreground capitalize">({detailedLog.actor.type})</span>
                        )}
                    </div>

                    {detailedLog.viaApiKey && (
                        <Badge
                            variant="outline"
                            className="text-[11px] gap-1 py-0 px-1.5 font-normal text-amber-600 dark:text-amber-400 border-amber-500/30 bg-amber-500/10"
                        >
                            <KeyRound className="size-3" />
                            API Key
                        </Badge>
                    )}

                    {ipAddress && (
                        <>
                            <span className="text-muted-foreground/40 select-none">•</span>
                            <div className="flex items-center gap-1.5">
                                <Globe className="size-3.5 shrink-0 text-muted-foreground/70" />
                                <span className="font-mono text-xs text-foreground/80">{ipAddress}</span>
                            </div>
                        </>
                    )}
                </div>

                {/* Line 3: Scope (if present for current scope context) */}
                {showSummaryScope && (
                    <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                        <div className="flex items-center gap-1.5">
                            <FolderTree className="size-3.5 shrink-0 text-muted-foreground/70" />
                            <span>Scope:</span>
                        </div>

                        <div className="flex flex-wrap items-center gap-2">
                            {!isProjectOrEnvScope && detailedLog.scopeProject && (
                                <div className="flex items-center gap-1.5">
                                    <Avatar
                                        name={detailedLog.scopeProject.name}
                                        src={detailedLog.scopeProject.photo}
                                        className="size-4 rounded-xs text-[9px] border border-border shrink-0"
                                    />
                                    <span className="font-medium text-foreground">{detailedLog.scopeProject.name}</span>
                                </div>
                            )}

                            {!isProjectOrEnvScope && detailedLog.scopeProject && detailedLog.scopeApp && (
                                <span className="text-muted-foreground/40 select-none">•</span>
                            )}

                            {detailedLog.scopeApp && (
                                <div className="flex items-center gap-1.5">
                                    <Avatar
                                        name={detailedLog.scopeApp.name}
                                        src={detailedLog.scopeApp.photo}
                                        className="size-4 rounded-xs text-[9px] border border-border shrink-0"
                                    />
                                    <span className="font-medium text-foreground">{detailedLog.scopeApp.name}</span>
                                    <span className="font-medium text-muted-foreground">
                                        ({detailedLog.scopeApp.env})
                                    </span>
                                </div>
                            )}

                            {!isProjectOrEnvScope &&
                                !detailedLog.scopeProject &&
                                !detailedLog.scopeApp &&
                                detailedLog.scopeUser && (
                                    <div className="flex items-center gap-1.5">
                                        <Avatar
                                            name={detailedLog.scopeUser.fullName ?? detailedLog.scopeUser.username}
                                            src={detailedLog.scopeUser.photo}
                                            className="size-4 rounded-full text-[9px] border border-border shrink-0"
                                        />
                                        <span className="font-medium text-foreground">
                                            {detailedLog.scopeUser.fullName ?? detailedLog.scopeUser.username}
                                        </span>
                                    </div>
                                )}
                        </div>
                    </div>
                )}

                {/* Line 4: Resource */}
                <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
                    <div className="flex items-center gap-1.5">
                        <Layers className="size-3.5 shrink-0 text-muted-foreground/70" />
                        <span>Resource:</span>
                    </div>

                    {detailedLog.resource ? (
                        <div className="flex flex-wrap items-center gap-1.5">
                            {detailedLog.resource.type && (
                                <Badge
                                    variant="secondary"
                                    className="text-xs font-normal capitalize px-1.5 py-0"
                                >
                                    {detailedLog.resource.type}
                                </Badge>
                            )}
                            <span className="font-medium text-foreground">{resourceName}</span>
                            {detailedLog.resource.id && detailedLog.resource.id !== resourceName && (
                                <span className="text-xs font-mono text-muted-foreground">
                                    ({detailedLog.resource.id})
                                </span>
                            )}
                        </div>
                    ) : (
                        <span className="text-muted-foreground/70 text-xs italic">Global / None</span>
                    )}
                </div>
            </div>

            {/* Expanded Details Section */}
            {isExpanded && (
                <div className="flex flex-col gap-3.5 pt-3.5 mt-2.5 border-t border-border/70">
                    {isFetchingDetail && !detailResponse && !cachedDetail ? (
                        <div className="flex items-center justify-center gap-2 py-6 text-sm text-muted-foreground">
                            <Loader2 className="size-4 animate-spin text-muted-foreground" />
                            <span>Loading details...</span>
                        </div>
                    ) : (
                        <>
                            {/* Scope Hierarchy Banner */}
                            {hasAnyScope && (
                                <div className="flex flex-col gap-1.5">
                                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                                        Scope Hierarchy
                                    </span>
                                    <div className="flex flex-wrap items-center gap-3 rounded-md border border-border bg-muted/30 px-3.5 py-2 text-xs">
                                        {detailedLog.scopeProject && (
                                            <div className="flex items-center gap-1.5">
                                                <span className="text-muted-foreground">Project:</span>
                                                <Avatar
                                                    name={detailedLog.scopeProject.name}
                                                    src={detailedLog.scopeProject.photo}
                                                    className="size-4.5 rounded-xs text-[9px] border border-border shrink-0"
                                                />
                                                <span className="font-medium text-foreground">
                                                    {detailedLog.scopeProject.name}
                                                </span>
                                            </div>
                                        )}

                                        {detailedLog.scopeApp && (
                                            <>
                                                {detailedLog.scopeProject && (
                                                    <span className="text-muted-foreground/40 select-none">•</span>
                                                )}
                                                <div className="flex items-center gap-1.5">
                                                    <span className="text-muted-foreground">App:</span>
                                                    <Avatar
                                                        name={detailedLog.scopeApp.name}
                                                        src={detailedLog.scopeApp.photo}
                                                        className="size-4.5 rounded-xs text-[9px] border border-border shrink-0"
                                                    />
                                                    <span className="font-medium text-foreground">
                                                        {detailedLog.scopeApp.name}
                                                    </span>
                                                    <span className="font-medium text-muted-foreground">
                                                        ({detailedLog.scopeApp.env})
                                                    </span>
                                                </div>
                                            </>
                                        )}

                                        {detailedLog.scopeUser && (
                                            <>
                                                {Boolean(detailedLog.scopeProject ?? detailedLog.scopeApp) && (
                                                    <span className="text-muted-foreground/40 select-none">•</span>
                                                )}
                                                <div className="flex items-center gap-1.5">
                                                    <span className="text-muted-foreground">User:</span>
                                                    <Avatar
                                                        name={
                                                            detailedLog.scopeUser.fullName ??
                                                            detailedLog.scopeUser.username
                                                        }
                                                        src={detailedLog.scopeUser.photo}
                                                        className="size-4.5 rounded-full text-[9px] border border-border shrink-0"
                                                    />
                                                    <span className="font-medium text-foreground">
                                                        {detailedLog.scopeUser.fullName ??
                                                            detailedLog.scopeUser.username}
                                                    </span>
                                                    {detailedLog.scopeUser.email && (
                                                        <span className="text-muted-foreground">
                                                            ({detailedLog.scopeUser.email})
                                                        </span>
                                                    )}
                                                </div>
                                            </>
                                        )}
                                    </div>
                                </div>
                            )}

                            {/* Detailed Meta Grid */}
                            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-x-6 gap-y-2.5 text-xs">
                                <div>
                                    <span className="text-muted-foreground">Log ID:</span>{" "}
                                    <span className="font-mono text-foreground font-medium select-all">
                                        {detailedLog.id}
                                    </span>
                                </div>

                                {detailedLog.requestId && (
                                    <div>
                                        <span className="text-muted-foreground">Request ID:</span>{" "}
                                        <span className="font-mono text-foreground font-medium select-all">
                                            {detailedLog.requestId}
                                        </span>
                                    </div>
                                )}

                                {detailedLog.sessionUid && (
                                    <div>
                                        <span className="text-muted-foreground">Session UID:</span>{" "}
                                        <span className="font-mono text-foreground font-medium select-all">
                                            {detailedLog.sessionUid}
                                        </span>
                                    </div>
                                )}

                                <div>
                                    <span className="text-muted-foreground">Actor ID:</span>{" "}
                                    <span className="font-mono text-foreground select-all">
                                        {detailedLog.actor?.id ?? "—"}
                                    </span>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Actor Type:</span>{" "}
                                    <span className="text-foreground capitalize">{detailedLog.actor?.type ?? "—"}</span>
                                </div>

                                {detailedLog.actor?.loggedName &&
                                    detailedLog.actor.loggedName !== detailedLog.actor.name && (
                                        <div>
                                            <span className="text-muted-foreground">Logged Name:</span>{" "}
                                            <span className="text-foreground">{detailedLog.actor.loggedName}</span>
                                        </div>
                                    )}

                                <div>
                                    <span className="text-muted-foreground">Via API Key:</span>{" "}
                                    <span className="text-foreground">{detailedLog.viaApiKey ? "Yes" : "No"}</span>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Client IP:</span>{" "}
                                    <span className="font-mono text-foreground select-all">
                                        {detailedLog.clientIp ?? "—"}
                                    </span>
                                </div>

                                <div>
                                    <span className="text-muted-foreground">Remote Addr:</span>{" "}
                                    <span className="font-mono text-foreground select-all">
                                        {detailedLog.remoteAddr ?? "—"}
                                    </span>
                                </div>

                                {detailedLog.resource?.id && (
                                    <div>
                                        <span className="text-muted-foreground">Resource ID:</span>{" "}
                                        <span className="font-mono text-foreground select-all">
                                            {detailedLog.resource.id}
                                        </span>
                                    </div>
                                )}

                                {detailedLog.resource?.type && (
                                    <div>
                                        <span className="text-muted-foreground">Resource Type:</span>{" "}
                                        <span className="text-foreground capitalize">{detailedLog.resource.type}</span>
                                    </div>
                                )}

                                {detailedLog.resource?.loggedName &&
                                    detailedLog.resource.loggedName !== detailedLog.resource.name && (
                                        <div>
                                            <span className="text-muted-foreground">Resource Logged Name:</span>{" "}
                                            <span className="text-foreground">{detailedLog.resource.loggedName}</span>
                                        </div>
                                    )}
                            </div>

                            {detailedLog.userAgent && (
                                <div className="text-xs pt-1">
                                    <span className="text-muted-foreground">User Agent:</span>{" "}
                                    <span className="font-mono text-foreground/80 break-all select-all">
                                        {detailedLog.userAgent}
                                    </span>
                                </div>
                            )}

                            {/* Detail Field */}
                            {detailedLog.detail && (
                                <div className="flex flex-col gap-1.5 pt-2">
                                    <span className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                                        Detail Payload
                                    </span>
                                    <div className="rounded-md border border-border bg-muted/40 p-3 max-h-80 overflow-y-auto">
                                        <pre className="font-mono text-xs text-foreground/90 whitespace-pre-wrap break-all select-all">
                                            {formatDetail(detailedLog.detail)}
                                        </pre>
                                    </div>
                                </div>
                            )}
                        </>
                    )}
                </div>
            )}
        </div>
    );
}

export function AuditLogSummaryCardSkeleton() {
    return (
        <div
            className="rounded-[8px] border border-l-muted-foreground/30 bg-background shadow-xs p-3 sm:p-4"
            style={{ borderLeftWidth: 4 }}
        >
            <div className="flex flex-col gap-2.5">
                {/* Header Skeleton */}
                <div className="flex flex-wrap items-center justify-between gap-3">
                    <div className="flex items-center gap-2">
                        <Skeleton className="h-5 w-16 rounded-md" />
                        <Skeleton className="h-5 w-24 rounded-md" />
                        <Skeleton className="h-5 w-16 rounded-md" />
                    </div>
                    <Skeleton className="h-4 w-44 rounded-md" />
                </div>

                {/* Actor & IP Skeleton */}
                <div className="flex items-center gap-3 pt-2 border-t border-border/50">
                    <Skeleton className="size-5 rounded-full" />
                    <Skeleton className="h-4 w-28 rounded-md" />
                    <Skeleton className="size-3 rounded-full" />
                    <Skeleton className="h-4 w-32 rounded-md" />
                </div>

                {/* Resource Skeleton */}
                <div className="flex items-center gap-2">
                    <Skeleton className="h-4 w-16 rounded-md" />
                    <Skeleton className="h-4 w-20 rounded-md" />
                    <Skeleton className="h-4 w-36 rounded-md" />
                </div>
            </div>
        </div>
    );
}
