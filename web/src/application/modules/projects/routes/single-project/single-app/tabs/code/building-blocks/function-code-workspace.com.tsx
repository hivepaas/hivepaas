import { useState } from "react";

import { Rocket, Undo2 } from "lucide-react";
import { toast } from "sonner";
import { AppDeploymentSettingsCommands } from "~/projects/data";
import { type FunctionFile, type FunctionMethod } from "~/projects/domain";
import { FunctionCodeEditor } from "~/projects/module-shared/components";
import { functionSettingsToPayload, functionSourceToPayload } from "~/projects/module-shared/utils";

import { Button } from "@/components/ui";

import { FunctionFiles } from "./function-files.com";

/**
 * A function's inline code: its files, one in the editor, saved and deployed
 * together. A save sends the function's settings back with these files.
 */
export function FunctionCodeWorkspace({ projectId, env, appId, settings, readOnly, children }: Props) {
    const savedFiles = settings.functionSource.code.inline?.files ?? [];
    const [files, setFiles] = useState<FunctionFile[]>(savedFiles);
    const [selectedPath, setSelectedPath] = useState(savedFiles[0]?.path ?? "");
    const selected = files.find(file => file.path === selectedPath) ?? files[0];
    const isDirty = JSON.stringify(files) !== JSON.stringify(savedFiles);

    const { mutate: save, isPending } = AppDeploymentSettingsCommands.useUpdateOne({
        onSuccess: () => {
            toast.success("Code saved, deployment started");
        },
    });

    function handleChange(content: string) {
        if (!selected) {
            return;
        }
        setFiles(current => current.map(file => (file.path === selected.path ? { ...file, content } : file)));
    }

    function handleSave() {
        save({
            projectID: projectId,
            env,
            appID: appId,
            updateVer: settings.updateVer,
            payload: functionSettingsToPayload(settings, functionSourceToPayload(settings.functionSource, files)),
        });
    }

    return (
        <div className="flex flex-col gap-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
                <span className="text-sm text-muted-foreground">{isDirty ? "Changes not saved" : "No changes"}</span>
                {!readOnly ? (
                    <div className="flex gap-2">
                        <Button
                            type="button"
                            variant="outline"
                            disabled={!isDirty || isPending}
                            onClick={() => {
                                setFiles(savedFiles);
                            }}
                        >
                            <Undo2 /> Discard
                        </Button>
                        <Button
                            type="button"
                            disabled={!isDirty}
                            isLoading={isPending}
                            onClick={handleSave}
                        >
                            <Rocket /> Save & Deploy
                        </Button>
                    </div>
                ) : null}
            </div>
            <div className="flex flex-col gap-4 xl:flex-row">
                <div className="flex min-h-[800px] min-w-0 flex-1 flex-col gap-4 lg:flex-row">
                    <FunctionFiles
                        files={files}
                        selectedPath={selected?.path ?? ""}
                        readOnly={readOnly}
                        onSelect={setSelectedPath}
                        onAdd={path => {
                            setFiles(current => [...current, { path, content: "" }]);
                            setSelectedPath(path);
                        }}
                        onRemove={path => {
                            setFiles(current => current.filter(file => file.path !== path));
                        }}
                    />
                    {selected ? (
                        <FunctionCodeEditor
                            path={selected.path}
                            value={selected.content}
                            onChange={handleChange}
                            readOnly={readOnly}
                        />
                    ) : null}
                </div>
                {children?.(files, setFiles)}
            </div>
        </div>
    );
}

interface Props {
    projectId: string;
    env: string;
    appId: string;
    settings: FunctionMethod;
    readOnly: boolean;
    /** What sits beside the editor - the test panel - given the files as edited. */
    children?: (files: FunctionFile[], setFiles: (files: FunctionFile[]) => void) => React.ReactNode;
}
