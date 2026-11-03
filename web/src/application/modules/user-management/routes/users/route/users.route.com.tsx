import { listBox } from "@lib/styles";
import { cn } from "@lib/utils";

import { UsersTable } from "../building-blocks";

function View() {
    return (
        <div className={cn(listBox)}>
            <UsersTable />
        </div>
    );
}

export function UsersRoute() {
    return <View />;
}
