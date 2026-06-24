import { CreateNetworkFormRoute } from "~/cluster/module-shared/components";

export function NetworkCreateRoute() {
    return <CreateNetworkFormRoute scope={{ type: "cluster" }} />;
}
