import { UsersTable } from "../building-blocks";

function View() {
    return (
        <div className="bg-background rounded-lg p-4">
            <UsersTable />
        </div>
    );
}

export function UsersRoute() {
    return <View />;
}
