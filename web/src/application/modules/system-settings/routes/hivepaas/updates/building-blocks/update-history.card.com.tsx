import { cn } from "@lib/utils";
import { Link } from "react-router";
import { SystemTasksQueries } from "~/operations/data";
import { SystemTaskStatus } from "~/operations/domain";

import { DEFAULT_PAGINATED_DATA, ROUTE } from "@application/shared/constants";
import { timeAgoFormatter } from "@application/shared/utils/time-ago";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

/** Each update runs as a task: this is the list of them, each leading to its log. */
const SYSTEM_UPDATE_TASK = "task:system-update";

const STATUS_LABEL: Record<string, string> = {
    [SystemTaskStatus.NotStarted]: "Waiting",
    [SystemTaskStatus.InProgress]: "Running",
    [SystemTaskStatus.Canceled]: "Canceled",
    [SystemTaskStatus.Done]: "Done",
    [SystemTaskStatus.Failed]: "Failed",
};

export function UpdateHistoryCard() {
    const { data: { data: tasks } = DEFAULT_PAGINATED_DATA, isLoading } = SystemTasksQueries.useFindManyPaginated({
        type: [SYSTEM_UPDATE_TASK],
        pagination: { page: 1, size: 5 },
        sorting: [{ id: "createdAt", desc: true }],
    });

    return (
        <Card className="gap-0 py-0">
            <CardHeader className="flex items-center border-b px-5 py-4 [.border-b]:pb-4">
                <CardTitle className="text-[15px]">Update history</CardTitle>
            </CardHeader>
            <CardContent className="px-0">
                {!isLoading && tasks.length === 0 ? (
                    <p className="px-5 py-5 text-sm text-muted-foreground">HivePaaS has not been updated here yet.</p>
                ) : (
                    <ul>
                        {tasks.map(task => (
                            <li
                                key={task.id}
                                className="border-b border-border/60 last:border-b-0"
                            >
                                <Link
                                    to={ROUTE.operations.tasks.details.$route(task.id)}
                                    className="grid grid-cols-[minmax(0,1fr)_auto_auto] items-center gap-4 px-5 py-3 text-[13px] hover:bg-muted/40"
                                >
                                    <span className="truncate">System update</span>
                                    <span
                                        className={cn(
                                            task.status === SystemTaskStatus.Failed && "font-semibold text-destructive",
                                            task.status === SystemTaskStatus.Done &&
                                                "text-green-700 dark:text-green-500",
                                        )}
                                    >
                                        {STATUS_LABEL[task.status] ?? task.status}
                                    </span>
                                    <span className="w-28 text-right text-muted-foreground">
                                        {timeAgoFormatter.format(task.createdAt)}
                                    </span>
                                </Link>
                            </li>
                        ))}
                    </ul>
                )}
            </CardContent>
        </Card>
    );
}
