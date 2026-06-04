import { NetworkManagementTable } from "~/cluster/module-shared/components";

export function NetworksTable() {
    return <NetworkManagementTable scope={{ type: "cluster" }} />;
}
