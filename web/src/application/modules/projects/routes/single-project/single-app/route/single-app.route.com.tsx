import { useParams } from "react-router";

function View() {
    const { id, appId } = useParams<{ id: string; appId: string }>();

    return (
        <div className="flex flex-col gap-5">
            <h1 className="text-2xl font-bold">Single App View</h1>
            <p>Project ID: {id}</p>
            <p>App ID: {appId}</p>
        </div>
    );
}

export function SingleAppRoute() {
    return <View />;
}
