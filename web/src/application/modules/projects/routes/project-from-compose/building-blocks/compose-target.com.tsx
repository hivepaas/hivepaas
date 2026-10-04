import { FieldError, Input, Select, SelectContent, SelectItem, SelectTrigger } from "@components/ui";
import { PlusIcon } from "lucide-react";
import { type ProjectEnvEntity } from "~/projects/domain";
import { ProjectEnvBadge } from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

/** How many envs a project may have, as creating it allows. */
const MAX_PROJECT_ENVS = 10;

/** The select's value for a new env, which no env's name can be: names are lowercase. */
const NEW_ENV = "+NEW";

/** Where the services go in an existing project: one of its envs, or a new one. */
export interface ComposeTargetValue {
    /** The env's name: one the project has, or the new one's. */
    env: string;
    newEnv: boolean;
    color: string;
}

/**
 * The project the services are added to, and the env: one it has - the
 * header's, when one is chosen there - or a new one, named and coloured.
 */
export function ComposeTarget({ projectName, envs, value, onChange, envError }: Props) {
    const selected = envs.find(env => env.name === value.env);
    const canAdd = envs.length < MAX_PROJECT_ENVS;

    return (
        <>
            <InfoBlock
                titleWidth={220}
                title={<LabelWithInfo label="Project" />}
            >
                <p className="text-sm text-foreground">{projectName}</p>
            </InfoBlock>

            <InfoBlock
                titleWidth={220}
                title={
                    <LabelWithInfo
                        label="Environment"
                        content="Where every service becomes an app. Nothing the env has is changed: a service named as one of its apps waits for your choice below."
                    />
                }
            >
                <div className="flex w-full flex-col gap-2">
                    <Select
                        value={value.newEnv ? NEW_ENV : value.env}
                        onValueChange={next => {
                            if (next === NEW_ENV) {
                                onChange({ ...value, env: "", newEnv: true });
                            } else {
                                onChange({ ...value, env: next, newEnv: false });
                            }
                        }}
                    >
                        <SelectTrigger
                            aria-invalid={envError !== undefined && !value.newEnv}
                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                        >
                            {value.newEnv ? (
                                <span>A new environment</span>
                            ) : selected ? (
                                <ProjectEnvBadge
                                    name={selected.name}
                                    color={selected.color}
                                />
                            ) : (
                                <span className="text-muted-foreground">Select environment</span>
                            )}
                        </SelectTrigger>
                        <SelectContent>
                            {envs.map(env => (
                                <SelectItem
                                    key={env.name}
                                    value={env.name}
                                >
                                    <ProjectEnvBadge
                                        name={env.name}
                                        color={env.color}
                                    />
                                </SelectItem>
                            ))}
                            <SelectItem
                                value={NEW_ENV}
                                disabled={!canAdd}
                            >
                                <PlusIcon className="size-4" />
                                {canAdd ? "A new environment" : `A new environment (${MAX_PROJECT_ENVS} at most)`}
                            </SelectItem>
                        </SelectContent>
                    </Select>

                    {value.newEnv && (
                        <div className={`flex items-center gap-2 ${PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}`}>
                            <Input
                                value={value.env}
                                onChange={event => {
                                    onChange({ ...value, env: event.target.value.toLowerCase() });
                                }}
                                placeholder="staging"
                                aria-invalid={envError !== undefined}
                            />
                            <label
                                className="inline-flex size-9 shrink-0 cursor-pointer items-center justify-center rounded-md border border-border focus-within:ring-2 focus-within:ring-ring focus-within:ring-offset-2"
                                style={{ backgroundColor: value.color }}
                                aria-label="The new environment's color"
                            >
                                <input
                                    type="color"
                                    value={value.color}
                                    onChange={event => {
                                        onChange({ ...value, color: event.target.value });
                                    }}
                                    className="sr-only"
                                />
                            </label>
                        </div>
                    )}
                    <FieldError errors={[envError === undefined ? undefined : { message: envError }]} />
                </div>
            </InfoBlock>
        </>
    );
}

interface Props {
    projectName: string;
    envs: ProjectEnvEntity[];
    value: ComposeTargetValue;
    onChange: (value: ComposeTargetValue) => void;
    /** What the env was refused for. */
    envError?: string;
}
