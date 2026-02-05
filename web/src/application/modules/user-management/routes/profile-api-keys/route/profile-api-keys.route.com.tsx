import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { ApiKeysTable } from "../building-blocks";

export function ProfileApiKeysRoute() {
    return (
        <div className={cn(listBox)}>
            <ApiKeysTable />
        </div>
    );
}
