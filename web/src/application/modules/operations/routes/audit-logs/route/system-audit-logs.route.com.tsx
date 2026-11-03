import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";

import { AuditLogsList } from "../building-blocks";

export function SystemAuditLogsRoute() {
    return (
        <section className={cn(listBox)}>
            <AuditLogsList scope={{ type: "global" }} />
        </section>
    );
}
