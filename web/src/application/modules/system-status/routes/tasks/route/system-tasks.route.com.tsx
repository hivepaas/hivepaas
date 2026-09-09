import { cn } from "@/lib/utils";
import { listBox } from "@lib/styles";

import { SystemTasksList } from "../building-blocks";

export function SystemTasksRoute() {
    return (
        <section className={cn(listBox)}>
            <SystemTasksList scope={{ type: "global" }} />
        </section>
    );
}
