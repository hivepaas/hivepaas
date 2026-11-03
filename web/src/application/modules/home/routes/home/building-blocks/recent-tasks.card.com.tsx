import { cn } from "@lib/utils";
import { Link } from "react-router";
import { SystemTasksQueries } from "~/operations/data";
import { type SystemTask, SystemTaskStatus } from "~/operations/domain";

import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

const RECENT_COUNT = 6;

const STATUS_LABEL: Record<SystemTask["status"], string> = {
    [SystemTaskStatus.NotStarted]: "Waiting",
    [SystemTaskStatus.InProgress]: "Running",
    [SystemTaskStatus.Canceled]: "Canceled",
    [SystemTaskStatus.Done]: "Done",
    [SystemTaskStatus.Failed]: "Failed",
};

const STATUS_CLASS: Record<SystemTask["status"], string> = {
    [SystemTaskStatus.NotStarted]: "text-muted-foreground",
    [SystemTaskStatus.InProgress]: "text-sky-700 dark:text-sky-400",
    [SystemTaskStatus.Canceled]: "text-muted-foreground",
    [SystemTaskStatus.Done]: "text-green-700 dark:text-green-500",
    [SystemTaskStatus.Failed]: "text-destructive font-semibold",
};

/** What the task was for: the app, else the project, else the job. */
function taskSubject(task: SystemTask): string | undefined {
    if (task.scopeApp) {
        return task.scopeProject ? `${task.scopeProject.name} / ${task.scopeApp.name}` : task.scopeApp.name;
    }
    return task.scopeProject?.name ?? task.targetJob?.name;
}

/** The latest tasks across the system. Shown to whoever may read the system screens. */
export function RecentTasksCard() {
    const { data: { data: tasks } = DEFAULT_PAGINATED_DATA, isLoading } = SystemTasksQueries.useFindManyPaginated(
        {
            pagination: { page: 1, size: RECENT_COUNT },
            sorting: [{ id: "createdAt", desc: true }],
        },
        { refetchInterval: 30_000 },
    );

    return (
        <Card className="gap-0 py-0">
            <CardHeader className="flex flex-row items-center justify-between border-b px-5 py-4 [.border-b]:pb-4">
                <CardTitle className="text-[15px]">Recent activity</CardTitle>
                <Link
                    to={ROUTE.operations.tasks.$route}
                    className="text-[13px] font-medium underline-offset-4 hover:underline"
                >
                    All tasks
                </Link>
            </CardHeader>
            <CardContent className="px-0">
                {isLoading ? (
                    <div className="flex flex-col gap-3 p-5">
                        <Skeleton className="h-6 w-full" />
                        <Skeleton className="h-6 w-full" />
                        <Skeleton className="h-6 w-full" />
                    </div>
                ) : tasks.length === 0 ? (
                    <p className="px-5 py-6 text-sm text-muted-foreground">No tasks have run yet.</p>
                ) : (
                    <ul>
                        {tasks.map(task => {
                            const subject = taskSubject(task);
                            return (
                                <li
                                    key={task.id}
                                    className="border-b border-border/60 last:border-b-0"
                                >
                                    <Link
                                        to={ROUTE.operations.tasks.details.$route(task.id)}
                                        className="grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-4 px-5 py-2.5 text-[13px] hover:bg-muted/40"
                                    >
                                        <span className="min-w-0 truncate">
                                            <span className="font-mono text-xs">{task.type}</span>
                                            {subject && <span className="text-muted-foreground"> · {subject}</span>}
                                        </span>
                                        <span className={cn(STATUS_CLASS[task.status])}>
                                            {STATUS_LABEL[task.status]}
                                        </span>
                                        <span className="w-24 text-right text-muted-foreground">
                                            {timeAgoFormatter.format(task.createdAt)}
                                        </span>
                                    </Link>
                                </li>
                            );
                        })}
                    </ul>
                )}
            </CardContent>
        </Card>
    );
}
