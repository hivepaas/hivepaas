import { useParams } from "react-router";
import invariant from "tiny-invariant";

export function ProjectSettingsDummyPage({ title }: Props) {
    const { id: projectId } = useParams<{ id: string }>();

    invariant(projectId, "projectId must be defined");

    return (
        <div className="flex flex-col gap-2">
            <h1 className="text-lg font-semibold text-foreground">{title}</h1>
            <p className="text-sm text-muted-foreground">This page is under construction.</p>
        </div>
    );
}

interface Props {
    title: string;
}
