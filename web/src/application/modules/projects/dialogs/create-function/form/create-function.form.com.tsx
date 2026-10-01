import React, { useEffect, useMemo } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { FileCode, GitBranch } from "lucide-react";
import { useController, useForm } from "react-hook-form";
import { ProjectDomainSettingsQueries } from "~/projects/data/queries";
import { type ProjectEnvEntity } from "~/projects/domain";
import {
    GitCredentialCombobox,
    GitCredentialLinks,
    type OptionCard,
    OptionCardGroup,
    ProjectEnvBadge,
} from "~/projects/module-shared/components";
import { FUNCTION_TEMPLATES } from "~/projects/module-shared/constants";
import { ALL_FUNCTION_RUNTIMES, EFunctionRuntime, FUNCTION_RUNTIME_LABELS } from "~/projects/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field, FieldError, FieldGroup } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import {
    type CreateFunctionFormInput,
    type CreateFunctionFormOutput,
    EFunctionCodeSource,
    createCreateFunctionFormSchema,
    suggestFunctionDomain,
} from "../schemas";

const CODE_SOURCE_OPTIONS: OptionCard<EFunctionCodeSource>[] = [
    {
        value: EFunctionCodeSource.Template,
        label: "Write it here",
        description: "Start from the runtime's template, and edit the code in the function's Code tab.",
        icon: FileCode,
    },
    {
        value: EFunctionCodeSource.Repository,
        label: "From a repository",
        description: "Build the function from a git repository's code.",
        icon: GitBranch,
    },
];

export function CreateFunctionForm({
    projectId,
    envs,
    initialEnv,
    isPending,
    readOnly = false,
    onSubmit,
    onHasChanges,
}: Props) {
    const envNames = useMemo(() => envs.map(env => env.name), [envs]);
    const defaultEnv = useMemo(() => {
        if (initialEnv && envNames.includes(initialEnv)) {
            return initialEnv;
        }

        return envs[0]?.name ?? "";
    }, [envNames, envs, initialEnv]);
    const schema = useMemo(() => createCreateFunctionFormSchema(envNames), [envNames]);

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
        watch,
        setValue,
    } = useForm<CreateFunctionFormInput, unknown, CreateFunctionFormOutput>({
        defaultValues: {
            name: "",
            env: defaultEnv,
            runtime: EFunctionRuntime.Node24,
            codeSource: EFunctionCodeSource.Template,
            repoUrl: "",
            repoRef: "",
            dir: "",
            credentials: null,
            expose: false,
            domain: "",
        },
        resolver: zodResolver(schema),
        mode: "onSubmit",
    });

    useEffect(() => {
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, onHasChanges, readOnly]);

    const selectedEnvName = watch("env");
    const runtime = watch("runtime") as EFunctionRuntime;
    const codeSource = watch("codeSource");
    const selectedEnv = envs.find(env => env.name === selectedEnvName);

    useEffect(() => {
        const currentEnvExists = envs.some(env => env.name === selectedEnvName);

        if (!currentEnvExists && selectedEnvName !== defaultEnv) {
            setValue("env", defaultEnv, { shouldDirty: false });
        }
    }, [defaultEnv, envs, selectedEnvName, setValue]);

    const { field: name } = useController({ name: "name", control });
    const { field: env } = useController({ name: "env", control });
    const { field: runtimeField } = useController({ name: "runtime", control });
    const { field: codeSourceField } = useController({ name: "codeSource", control });
    const { field: repoUrl } = useController({ name: "repoUrl", control });
    const { field: repoRef } = useController({ name: "repoRef", control });
    const { field: dir } = useController({ name: "dir", control });
    const { field: credentials } = useController({ name: "credentials", control });
    const { field: expose } = useController({ name: "expose", control });
    const { field: domain } = useController({ name: "domain", control });

    // The project's root domain, under which a function is offered a domain.
    const { data: domainSettings } = ProjectDomainSettingsQueries.useFindOne({ projectID: projectId });
    const rootDomain = domainSettings?.data.rootDomain ?? "";

    function changeExpose(checked: boolean) {
        expose.onChange(checked);
        if (checked && domain.value === "") {
            setValue("domain", suggestFunctionDomain(watch("name"), rootDomain));
        }
    }

    function onValid(values: CreateFunctionFormOutput) {
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid)(event);
            }}
            className="min-h-0 flex flex-1 flex-col"
        >
            <DialogBody className="flex flex-col gap-6">
                <fieldset
                    disabled={readOnly}
                    className="m-0 flex min-w-0 flex-col gap-6 border-0 p-0"
                >
                    <InfoBlock
                        titleWidth={150}
                        title={
                            <LabelWithInfo
                                label="Name"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Input
                                    id="name"
                                    {...name}
                                    placeholder="E.g. hello"
                                    aria-invalid={Boolean(errors.name)}
                                />
                                <FieldError errors={[errors.name]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={150}
                        title={
                            <LabelWithInfo
                                label="Environment"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Select
                                    value={env.value}
                                    onValueChange={value => {
                                        env.onChange(value);
                                        // Credentials are an env's: another env lists its own.
                                        setValue("credentials", null);
                                    }}
                                    disabled={readOnly || envs.length === 0}
                                >
                                    <SelectTrigger
                                        aria-invalid={Boolean(errors.env)}
                                        className="px-2"
                                    >
                                        {selectedEnv ? (
                                            <ProjectEnvBadge
                                                name={selectedEnv.name}
                                                color={selectedEnv.color}
                                            />
                                        ) : (
                                            <span className="text-muted-foreground">
                                                {envs.length === 0 ? "No environments" : "Select environment"}
                                            </span>
                                        )}
                                    </SelectTrigger>
                                    <SelectContent>
                                        {envs.map(projectEnv => (
                                            <SelectItem
                                                key={projectEnv.name}
                                                value={projectEnv.name}
                                            >
                                                <ProjectEnvBadge
                                                    name={projectEnv.name}
                                                    color={projectEnv.color}
                                                />
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <FieldError errors={[errors.env]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={150}
                        title={
                            <LabelWithInfo
                                label="Runtime"
                                isRequired
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <Select
                                    value={runtimeField.value}
                                    onValueChange={runtimeField.onChange}
                                    disabled={readOnly}
                                >
                                    <SelectTrigger>
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent>
                                        {ALL_FUNCTION_RUNTIMES.map(item => (
                                            <SelectItem
                                                key={item}
                                                value={item}
                                            >
                                                {FUNCTION_RUNTIME_LABELS[item]}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <FieldError errors={[errors.runtime]} />
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    <InfoBlock
                        titleWidth={150}
                        title={<LabelWithInfo label="Code" />}
                    >
                        <FieldGroup>
                            <Field>
                                <OptionCardGroup
                                    options={CODE_SOURCE_OPTIONS}
                                    value={codeSourceField.value}
                                    onChange={codeSourceField.onChange}
                                    className="grid-cols-1 sm:grid-cols-2"
                                    readOnly={readOnly}
                                />
                                {codeSource === EFunctionCodeSource.Template && (
                                    <p className="text-xs text-muted-foreground">
                                        Starts with{" "}
                                        {FUNCTION_TEMPLATES[runtime].map(file => (
                                            <code
                                                key={file.path}
                                                className="mr-1 rounded bg-muted px-1"
                                            >
                                                {file.path}
                                            </code>
                                        ))}
                                    </p>
                                )}
                            </Field>
                        </FieldGroup>
                    </InfoBlock>

                    {codeSource === EFunctionCodeSource.Repository && (
                        <>
                            <InfoBlock
                                titleWidth={150}
                                title={
                                    <LabelWithInfo
                                        label="Repository URL"
                                        isRequired
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            {...repoUrl}
                                            placeholder="https://github.com/owner/repo.git"
                                            aria-invalid={Boolean(errors.repoUrl)}
                                        />
                                        <FieldError errors={[errors.repoUrl]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={150}
                                title={
                                    <LabelWithInfo
                                        label="Ref"
                                        content="A branch, a tag or a commit; the default branch when empty."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            {...repoRef}
                                            placeholder="main"
                                        />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={150}
                                title={
                                    <LabelWithInfo
                                        label="Directory"
                                        content="The function's directory in the repository; its root when empty."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            {...dir}
                                            placeholder="functions/hello"
                                            aria-invalid={Boolean(errors.dir)}
                                        />
                                        <FieldError errors={[errors.dir]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={150}
                                title={<LabelWithInfo label="Git Credentials" />}
                            >
                                <FieldGroup>
                                    <Field>
                                        <GitCredentialCombobox
                                            projectId={projectId}
                                            env={selectedEnvName}
                                            value={credentials.value}
                                            onChange={credentials.onChange}
                                            readOnly={readOnly}
                                        />
                                        <GitCredentialLinks projectId={projectId} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>
                        </>
                    )}

                    <InfoBlock
                        titleWidth={150}
                        title={
                            <LabelWithInfo
                                label="Domain"
                                content="Off: the function is reached in its project, by its name, and by its scheduled calls."
                            />
                        }
                    >
                        <FieldGroup>
                            <Field>
                                <div className="flex items-center gap-2 text-sm">
                                    <Checkbox
                                        id="create-function-expose"
                                        checked={expose.value}
                                        onCheckedChange={checked => {
                                            changeExpose(checked === true);
                                        }}
                                        disabled={readOnly}
                                    />
                                    <label htmlFor="create-function-expose">Expose at a domain</label>
                                </div>
                                {expose.value && (
                                    <>
                                        <Input
                                            {...domain}
                                            placeholder="hello.example.com"
                                            aria-invalid={Boolean(errors.domain)}
                                        />
                                        <FieldError errors={[errors.domain]} />
                                        <p className="text-xs text-muted-foreground">
                                            Routed over HTTPS, forced; turn it off in the routing settings. A
                                            certificate that covers the domain is attached, or obtained when the project
                                            obtains them; otherwise the proxy&apos;s own answers.
                                        </p>
                                    </>
                                )}
                            </Field>
                        </FieldGroup>
                    </InfoBlock>
                </fieldset>
            </DialogBody>
            <DialogActionFooter>
                <Button
                    type="submit"
                    isLoading={isPending}
                    disabled={readOnly}
                >
                    Create Function
                </Button>
            </DialogActionFooter>
        </form>
    );
}

interface Props {
    projectId: string;
    envs: ProjectEnvEntity[];
    initialEnv?: string;
    isPending: boolean;
    readOnly?: boolean;
    onSubmit: (values: CreateFunctionFormOutput) => Promise<void> | void;
    onHasChanges?: (dirty: boolean) => void;
}
