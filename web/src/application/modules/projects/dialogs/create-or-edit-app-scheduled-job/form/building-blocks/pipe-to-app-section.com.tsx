import { useEffect, useState } from "react";

import { useController, useFormContext } from "react-hook-form";
import { AppsPickerDialog, CommandConfigSection, ProjectsPickerDialog } from "~/projects/module-shared/components";

import { ContentBlock, InfoBlock, LabelWithInfo } from "@application/shared/components";
import { ROUTE } from "@application/shared/constants";

import { Button, Field, FieldError, FieldGroup, Input } from "@/components/ui";

import type { CreateOrEditAppScheduledJobFormInput } from "../../schemas";

type SchemaInput = CreateOrEditAppScheduledJobFormInput;

interface Props {
    projectId: string;
    appId: string;
    readOnly?: boolean;
}

export function PipeToAppSection({ projectId, appId, readOnly = false }: Props) {
    const [isProjectsDialogOpen, setProjectsDialogOpen] = useState(false);
    const [isAppsDialogOpen, setAppsDialogOpen] = useState(false);

    const { control, setValue } = useFormContext<SchemaInput>();

    const { field: pipeTargetProject } = useController({ control, name: "pipeTargetProject" });
    const {
        field: pipeTargetApp,
        fieldState: { invalid: isPipeTargetAppInvalid, error: pipeTargetAppError },
    } = useController({ control, name: "pipeTargetApp" });

    const selectedProjectId = pipeTargetProject.value?.id ?? projectId;

    useEffect(() => {
        if (!pipeTargetProject.value) {
            pipeTargetProject.onChange({ id: projectId, name: "" });
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [projectId]);

    return (
        <>
            <ContentBlock label="Command Output: Pipe to App">
                <div className="flex flex-col gap-6">
                    <InfoBlock
                        title={<LabelWithInfo label="Target Project" />}
                        titleWidth={160}
                    >
                        <FieldGroup>
                            <Field>
                                <div className="flex flex-wrap items-start gap-3">
                                    <Input
                                        value={pipeTargetProject.value?.name ?? ""}
                                        placeholder="Select target project"
                                        readOnly
                                        className="w-full min-w-[260px] max-w-[400px] flex-1"
                                        disabled={readOnly}
                                    />
                                    {!readOnly && (
                                        <Button
                                            type="button"
                                            variant="outline"
                                            className="min-w-[130px]"
                                            onClick={() => {
                                                setProjectsDialogOpen(true);
                                            }}
                                        >
                                            Show Projects
                                        </Button>
                                    )}
                                </div>
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        title={
                            <LabelWithInfo
                                label="Target App"
                                isRequired
                            />
                        }
                        titleWidth={160}
                    >
                        <FieldGroup>
                            <Field>
                                <div className="flex flex-wrap items-start gap-3">
                                    <Input
                                        value={pipeTargetApp.value?.name ?? ""}
                                        placeholder="Select target app"
                                        readOnly
                                        aria-invalid={isPipeTargetAppInvalid}
                                        className="w-full min-w-[260px] max-w-[400px] flex-1"
                                        disabled={readOnly}
                                    />
                                    {!readOnly && (
                                        <Button
                                            type="button"
                                            variant="outline"
                                            className="min-w-[130px]"
                                            onClick={() => {
                                                setAppsDialogOpen(true);
                                            }}
                                        >
                                            Show Apps
                                        </Button>
                                    )}
                                </div>
                                <FieldError errors={[pipeTargetAppError]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>
                </div>
            </ContentBlock>

            <ProjectsPickerDialog
                open={isProjectsDialogOpen}
                onOpenChange={setProjectsDialogOpen}
                onSelect={project => {
                    pipeTargetProject.onChange({ id: project.id, name: project.name });
                    setValue("pipeTargetApp", null);
                }}
            />

            <AppsPickerDialog
                open={isAppsDialogOpen}
                onOpenChange={setAppsDialogOpen}
                projectID={selectedProjectId}
                excludeAppId={selectedProjectId === projectId ? appId : undefined}
                onSelect={app => {
                    pipeTargetApp.onChange({ id: app.id, name: app.name });
                }}
            />
        </>
    );
}

interface CommandRunInTargetAppProps {
    projectId: string;
    appId: string;
    readOnly?: boolean;
}

export function CommandRunInTargetApp({ projectId, appId: _appId, readOnly = false }: CommandRunInTargetAppProps) {
    const { watch } = useFormContext<SchemaInput>();
    const pipeTargetProject = watch("pipeTargetProject");
    const selectedProjectId = pipeTargetProject?.id ?? projectId;

    const configureTemplatesLink =
        ROUTE.projects.single.providerConfiguration.commandTemplates.$route(selectedProjectId);

    return (
        <CommandConfigSection
            label="Command Run in Target App"
            fieldPrefix="pipeCommand"
            showLoadTemplate
            templateProjectId={selectedProjectId}
            configureTemplatesLink={configureTemplatesLink}
            readOnly={readOnly}
            showArgGroups
            envLabel="Env"
        />
    );
}
