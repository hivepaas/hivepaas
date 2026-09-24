import { Checkbox, FieldError, Input } from "@components/ui";
import { InputNumber } from "@components/ui/input-number";
import { useController, useFormContext } from "react-hook-form";
import { type AppDockerApiLimits } from "~/projects/domain";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";
import { SingleValueList } from "@application/shared/form";

import {
    type AppConfigDockerApiFormSchemaInput,
    type AppConfigDockerApiFormSchemaOutput,
    DOCKER_API_GROUPS,
    DOCKER_API_MAX_CONTAINERS,
} from "../schemas";

const TITLE_WIDTH = 220;

/** What the app's containers may run and share, through the proxy. */
export function DockerApiPolicyFields({ readOnly = false }: { readOnly?: boolean }) {
    const {
        control,
        formState: { errors },
    } = useFormContext<AppConfigDockerApiFormSchemaInput, unknown, AppConfigDockerApiFormSchemaOutput>();
    const { field: envNetworkField } = useController({ control, name: "envNetwork" });
    const { field: allowField } = useController({ control, name: "allow" });

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                titleWidth={TITLE_WIDTH}
                title={
                    <LabelWithInfo
                        label="Images"
                        content='Patterns over the images its containers may run, such as autobase/automation or ghcr.io/org/*. "*" is any image.'
                    />
                }
            >
                <div className="flex flex-col gap-1">
                    <SingleValueList<AppConfigDockerApiFormSchemaInput>
                        name="images"
                        placeholder="autobase/automation"
                        className="max-w-[590px]"
                        enableValueEditing
                        disabled={readOnly}
                    />
                    {errors.images?.message && <FieldError>{errors.images.message}</FieldError>}
                </div>
            </InfoBlock>

            <InfoBlock
                titleWidth={TITLE_WIDTH}
                title={
                    <LabelWithInfo
                        label="Shared directories"
                        content="Directories of the app its containers may bind, at most five. Each must be on a volume the app mounts in Persistent Storage."
                    />
                }
            >
                <SingleValueList<AppConfigDockerApiFormSchemaInput>
                    name="sharedDirs"
                    placeholder="/var/lib/app/work"
                    className="max-w-[590px]"
                    enableValueEditing
                    disabled={readOnly}
                />
            </InfoBlock>

            <InfoBlock
                titleWidth={TITLE_WIDTH}
                title={
                    <LabelWithInfo
                        label="Env network"
                        content="Its containers may also join the env's network, to reach the other apps of the env - a CI job cloning from a Git server beside it."
                    />
                }
            >
                <Checkbox
                    checked={envNetworkField.value}
                    onCheckedChange={value => {
                        envNetworkField.onChange(value === true);
                    }}
                />
            </InfoBlock>

            <InfoBlock
                titleWidth={TITLE_WIDTH}
                title={
                    <LabelWithInfo
                        label="Allowed"
                        content="What its containers may do beyond being started, watched and removed."
                    />
                }
            >
                <div className="flex flex-col gap-2.5">
                    {DOCKER_API_GROUPS.map(group => {
                        const checked = allowField.value.includes(group.value);
                        const id = `docker-api-allow-${group.value}`;
                        return (
                            <div
                                key={group.value}
                                className="flex items-start gap-2.5 text-sm"
                            >
                                <Checkbox
                                    id={id}
                                    checked={checked}
                                    onCheckedChange={value => {
                                        allowField.onChange(
                                            value === true
                                                ? [...allowField.value, group.value]
                                                : allowField.value.filter(item => item !== group.value),
                                        );
                                    }}
                                    className="mt-0.5"
                                />
                                <label htmlFor={id}>
                                    <span className="font-medium">{group.label}</span>
                                    <span className="text-muted-foreground"> - {group.description}</span>
                                </label>
                            </div>
                        );
                    })}
                </div>
            </InfoBlock>
        </div>
    );
}

/** What each of the app's containers may use, and how many there may be. Empty keeps HivePaaS's default. */
export function DockerApiLimitFields({ defaults }: { defaults: AppDockerApiLimits }) {
    const { control } = useFormContext<
        AppConfigDockerApiFormSchemaInput,
        unknown,
        AppConfigDockerApiFormSchemaOutput
    >();
    const { field: containersField } = useController({ control, name: "limits.containers" });
    const { field: memoryField } = useController({ control, name: "limits.memory" });
    const { field: cpusField } = useController({ control, name: "limits.cpus" });

    return (
        <div className="flex flex-col gap-6">
            <InfoBlock
                titleWidth={TITLE_WIDTH}
                title={
                    <LabelWithInfo
                        label="Containers"
                        content={`How many containers the app may have at once, running or not. At most ${DOCKER_API_MAX_CONTAINERS}.`}
                    />
                }
            >
                <InputNumber
                    value={containersField.value}
                    onValueChange={value => {
                        containersField.onChange(value);
                    }}
                    placeholder={String(defaults.containers)}
                    className="max-w-[100px]"
                    min={0}
                    max={DOCKER_API_MAX_CONTAINERS}
                />
            </InfoBlock>

            <InfoBlock
                titleWidth={TITLE_WIDTH}
                title={
                    <LabelWithInfo
                        label="Memory"
                        content="The most memory one container may have, and what it gets when it asks for none. Use sizes like 512mb or 2gb."
                    />
                }
            >
                <Input
                    value={memoryField.value ?? ""}
                    onChange={memoryField.onChange}
                    placeholder={defaults.memory}
                    className="max-w-[100px]"
                />
            </InfoBlock>

            <InfoBlock
                titleWidth={TITLE_WIDTH}
                title={
                    <LabelWithInfo
                        label="CPUs"
                        content="The most processor time one container may have, and what it gets when it asks for none."
                    />
                }
            >
                <InputNumber
                    value={cpusField.value}
                    onValueChange={value => {
                        cpusField.onChange(value);
                    }}
                    placeholder={String(defaults.cpus)}
                    className="max-w-[100px]"
                    decimalScale={2}
                    fixedDecimalScale={false}
                    stepper={0.25}
                    min={0}
                />
            </InfoBlock>
        </div>
    );
}
