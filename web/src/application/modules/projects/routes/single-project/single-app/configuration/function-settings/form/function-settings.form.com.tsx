import React, { type PropsWithChildren, useImperativeHandle } from "react";

import { Checkbox, Field, FieldError, FieldGroup, Input } from "@components/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { FileCode, GitBranch } from "lucide-react";
import { useController, useForm } from "react-hook-form";
import { type FunctionSource } from "~/projects/domain";
import {
    GitCredentialCombobox,
    GitCredentialLinks,
    GitRepositoryUrlInput,
    type OptionCard,
    OptionCardGroup,
    PushToRegistryCombobox,
    RegistryCredentialsLink,
} from "~/projects/module-shared/components";
import { PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/projects/module-shared/constants";
import {
    ALL_FUNCTION_RUNTIMES,
    EFunctionRuntime,
    FUNCTION_JAVASCRIPT_RUNTIMES,
    functionRuntimeDefaultEntrypoint,
    functionRuntimeLabel,
    isKnownFunctionRuntime,
} from "~/projects/module-shared/enums";

import { ContentBlock, InfoBlock, LabelWithInfo } from "@application/shared/components";

import { type ValidationException } from "@infrastructure/exceptions/validation";

import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { TagInput } from "@/components/ui/tag-input";

import {
    EFunctionCodeLocation,
    type FunctionSettingsFormInput,
    type FunctionSettingsFormOutput,
    FunctionSettingsFormSchema,
    functionSettingsDefaultValues,
    functionSettingsErrorField,
} from "../schemas";

const CODE_LOCATION_OPTIONS: OptionCard<EFunctionCodeLocation>[] = [
    {
        value: EFunctionCodeLocation.Inline,
        label: "Written here",
        description: "The code is edited in the Code tab.",
        icon: FileCode,
    },
    {
        value: EFunctionCodeLocation.Repository,
        label: "From a repository",
        description: "The function is built from a git repository's code.",
        icon: GitBranch,
    },
];

const TITLE_WIDTH = 220;

export interface FunctionSettingsFormRef {
    /** Puts a backend's validation errors on the fields they are about. */
    onError: (error: ValidationException) => void;
}

export function FunctionSettingsForm({ ref, projectId, env, source, onSubmit, readOnly = false, children }: Props) {
    const {
        control,
        handleSubmit,
        watch,
        getValues,
        setValue,
        setError,
        formState: { errors },
    } = useForm<FunctionSettingsFormInput, unknown, FunctionSettingsFormOutput>({
        defaultValues: functionSettingsDefaultValues(source),
        resolver: zodResolver(FunctionSettingsFormSchema),
        mode: "onSubmit",
    });

    useImperativeHandle(
        ref,
        () => ({
            onError(error: ValidationException) {
                error.errors.forEach(({ path, message }) => {
                    const field = functionSettingsErrorField(path);
                    if (field) {
                        setError(field, { message, type: "manual" });
                    }
                });
            },
        }),
        [setError],
    );

    const runtime = watch("runtime");
    const codeLocation = watch("codeLocation");

    const { field: runtimeField } = useController({ control, name: "runtime" });
    const { field: entrypointFile } = useController({ control, name: "entrypointFile" });
    const { field: entrypointHandler } = useController({ control, name: "entrypointHandler" });
    const { field: timeout } = useController({ control, name: "timeout" });
    const { field: maxConcurrency } = useController({ control, name: "maxConcurrency" });
    const { field: maxBodySize } = useController({ control, name: "maxBodySize" });
    const { field: systemPackages } = useController({ control, name: "systemPackages" });
    const { field: codeLocationField } = useController({ control, name: "codeLocation" });
    const { field: repoUrl } = useController({ control, name: "repoUrl" });
    const { field: repoRef } = useController({ control, name: "repoRef" });
    const { field: dir } = useController({ control, name: "dir" });
    const { field: credentials } = useController({ control, name: "credentials" });
    const { field: autoDeploy } = useController({ control, name: "autoDeploy" });
    const { field: pushToRegistry } = useController({ control, name: "pushToRegistry" });

    // An entrypoint left at its runtime's default follows the runtime.
    function changeRuntime(next: string) {
        const previous = functionRuntimeDefaultEntrypoint(getValues("runtime"));
        const nextDefault = functionRuntimeDefaultEntrypoint(next);
        if (nextDefault && getValues("entrypointFile") === previous?.file) {
            setValue("entrypointFile", nextDefault.file, { shouldDirty: true });
        }
        if (nextDefault && getValues("entrypointHandler") === previous?.handler) {
            setValue("entrypointHandler", nextDefault.handler, { shouldDirty: true });
        }
        runtimeField.onChange(next);
    }

    function onValid(values: FunctionSettingsFormOutput) {
        if (readOnly) {
            return;
        }

        onSubmit(values);
    }

    const defaultEntrypoint = functionRuntimeDefaultEntrypoint(runtime);
    // A runtime the dashboard does not know stays offered, so the select shows it.
    const runtimes: string[] = isKnownFunctionRuntime(source.runtime)
        ? ALL_FUNCTION_RUNTIMES
        : [...ALL_FUNCTION_RUNTIMES, source.runtime];

    return (
        <div className="pt-2">
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid)(event);
                }}
                className="flex flex-col gap-6"
            >
                <fieldset
                    disabled={readOnly}
                    className="contents"
                >
                    <ContentBlock label="Runtime">
                        <div className="flex flex-col gap-6">
                            <InfoBlock
                                titleWidth={TITLE_WIDTH}
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
                                            onValueChange={changeRuntime}
                                            disabled={readOnly}
                                        >
                                            <SelectTrigger className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                                <SelectValue />
                                            </SelectTrigger>
                                            <SelectContent>
                                                {runtimes.map(item => (
                                                    <SelectItem
                                                        key={item}
                                                        value={item}
                                                    >
                                                        {functionRuntimeLabel(item)}
                                                    </SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                        <FieldError errors={[errors.runtime]} />
                                        {runtime !== source.runtime &&
                                            !(
                                                FUNCTION_JAVASCRIPT_RUNTIMES.includes(runtime as EFunctionRuntime) &&
                                                FUNCTION_JAVASCRIPT_RUNTIMES.includes(
                                                    source.runtime as EFunctionRuntime,
                                                )
                                            ) &&
                                            codeLocation === EFunctionCodeLocation.Inline && (
                                                <p className="text-xs text-muted-foreground">
                                                    The code stays as it is: rewrite it for the new runtime in the Code
                                                    tab.
                                                </p>
                                            )}
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={TITLE_WIDTH}
                                title={
                                    <LabelWithInfo
                                        label="Entrypoint File"
                                        content="The file of the handler; for Go, the directory of its package."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            {...entrypointFile}
                                            placeholder={defaultEntrypoint?.file}
                                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                            aria-invalid={Boolean(errors.entrypointFile)}
                                        />
                                        <FieldError errors={[errors.entrypointFile]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={TITLE_WIDTH}
                                title={
                                    <LabelWithInfo
                                        label="Handler"
                                        content="The name the file exports the handler by."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            {...entrypointHandler}
                                            placeholder={defaultEntrypoint?.handler}
                                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                            aria-invalid={Boolean(errors.entrypointHandler)}
                                        />
                                        <FieldError errors={[errors.entrypointHandler]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={TITLE_WIDTH}
                                title={
                                    <LabelWithInfo
                                        label="Debian Packages"
                                        content="Installed into the function's image, e.g. ffmpeg or libvips=8.14.1-3."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <TagInput
                                            tags={systemPackages.value}
                                            onCreate={pkg => {
                                                const name = pkg.trim();
                                                if (!readOnly && name !== "" && !systemPackages.value.includes(name)) {
                                                    systemPackages.onChange([...systemPackages.value, name]);
                                                }
                                            }}
                                            onDelete={pkg => {
                                                if (!readOnly) {
                                                    systemPackages.onChange(
                                                        systemPackages.value.filter(item => item !== pkg),
                                                    );
                                                }
                                            }}
                                            placeholder="Enter a package"
                                            disabled={readOnly}
                                        />
                                        <FieldError errors={[errors.systemPackages]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>
                        </div>
                    </ContentBlock>

                    <ContentBlock label="Limits Of A Call">
                        <div className="flex flex-col gap-6">
                            <InfoBlock
                                titleWidth={TITLE_WIDTH}
                                title={
                                    <LabelWithInfo
                                        label="Timeout"
                                        isRequired
                                        content="How long a call may take, from 1s to 15m, e.g. 30s or 2m."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            {...timeout}
                                            placeholder="30s"
                                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                            aria-invalid={Boolean(errors.timeout)}
                                        />
                                        <FieldError errors={[errors.timeout]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={TITLE_WIDTH}
                                title={
                                    <LabelWithInfo
                                        label="Concurrency"
                                        isRequired
                                        content="How many calls one instance runs at once; a call over it is answered 429."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            type="number"
                                            min={1}
                                            max={1000}
                                            name={maxConcurrency.name}
                                            ref={maxConcurrency.ref}
                                            value={String(maxConcurrency.value)}
                                            onChange={maxConcurrency.onChange}
                                            onBlur={maxConcurrency.onBlur}
                                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                            aria-invalid={Boolean(errors.maxConcurrency)}
                                        />
                                        <FieldError errors={[errors.maxConcurrency]} />
                                        {runtime === EFunctionRuntime.Python313 && (
                                            <p className="text-xs text-muted-foreground">
                                                Python serves with a process per CPU, at most this many and one per 128
                                                MiB of a memory limit. Set WEB_CONCURRENCY in the function&apos;s
                                                environment to choose.
                                            </p>
                                        )}
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>

                            <InfoBlock
                                titleWidth={TITLE_WIDTH}
                                title={
                                    <LabelWithInfo
                                        label="Body Size"
                                        isRequired
                                        content="The largest request body a call takes, from 1kb to 100mb, e.g. 6mb."
                                    />
                                }
                            >
                                <FieldGroup>
                                    <Field>
                                        <Input
                                            {...maxBodySize}
                                            placeholder="6mb"
                                            className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                            aria-invalid={Boolean(errors.maxBodySize)}
                                        />
                                        <FieldError errors={[errors.maxBodySize]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>
                        </div>
                    </ContentBlock>

                    <ContentBlock label="Code">
                        <div className="flex flex-col gap-6">
                            <OptionCardGroup
                                options={CODE_LOCATION_OPTIONS}
                                value={codeLocationField.value}
                                onChange={codeLocationField.onChange}
                                className="grid-cols-1 sm:grid-cols-2 max-w-[600px]"
                                readOnly={readOnly}
                            />

                            {codeLocation === EFunctionCodeLocation.Inline && source.code.repo && (
                                <p className="text-xs text-muted-foreground">
                                    The code starts from the runtime&apos;s template, to edit in the Code tab.
                                </p>
                            )}

                            {codeLocation === EFunctionCodeLocation.Repository && (
                                <>
                                    {source.code.inline && (
                                        <p className="text-xs text-muted-foreground">
                                            Saving drops the code written here: the function is built from the
                                            repository.
                                        </p>
                                    )}
                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
                                        title="Git Credentials"
                                    >
                                        <FieldGroup>
                                            <Field>
                                                <GitCredentialCombobox
                                                    projectId={projectId}
                                                    env={env}
                                                    value={credentials.value ?? null}
                                                    onChange={credentials.onChange}
                                                    readOnly={readOnly}
                                                    invalid={Boolean(errors.credentials)}
                                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                />
                                                <FieldError errors={[errors.credentials]} />
                                                <GitCredentialLinks projectId={projectId} />
                                            </Field>
                                        </FieldGroup>
                                    </InfoBlock>

                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
                                        title={
                                            <LabelWithInfo
                                                label="Repository URL"
                                                isRequired
                                            />
                                        }
                                    >
                                        <FieldGroup>
                                            <Field>
                                                <div className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                                    <GitRepositoryUrlInput
                                                        projectId={projectId}
                                                        env={env}
                                                        credentials={credentials.value}
                                                        value={repoUrl.value}
                                                        onChange={repoUrl.onChange}
                                                        invalid={Boolean(errors.repoUrl)}
                                                        readOnly={readOnly}
                                                    />
                                                </div>
                                                <FieldError errors={[errors.repoUrl]} />
                                            </Field>
                                        </FieldGroup>
                                    </InfoBlock>

                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
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
                                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                    aria-invalid={Boolean(errors.repoRef)}
                                                />
                                                <FieldError errors={[errors.repoRef]} />
                                            </Field>
                                        </FieldGroup>
                                    </InfoBlock>

                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
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
                                                    className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                                    aria-invalid={Boolean(errors.dir)}
                                                />
                                                <FieldError errors={[errors.dir]} />
                                            </Field>
                                        </FieldGroup>
                                    </InfoBlock>

                                    <InfoBlock
                                        titleWidth={TITLE_WIDTH}
                                        title={
                                            <LabelWithInfo
                                                label="Deploy on Push"
                                                content="A push to the ref deploys the function, once the repository's webhook or GitHub App tells HivePaaS of it."
                                            />
                                        }
                                    >
                                        <FieldGroup>
                                            <Field>
                                                <div className="flex items-center gap-3">
                                                    <Checkbox
                                                        id="function-auto-deploy"
                                                        checked={autoDeploy.value}
                                                        onCheckedChange={checked => {
                                                            autoDeploy.onChange(checked === true);
                                                        }}
                                                        disabled={readOnly}
                                                    />
                                                    <label
                                                        htmlFor="function-auto-deploy"
                                                        className="text-sm"
                                                    >
                                                        Deploy when the ref is pushed
                                                    </label>
                                                </div>
                                            </Field>
                                        </FieldGroup>
                                    </InfoBlock>
                                </>
                            )}
                        </div>
                    </ContentBlock>

                    <ContentBlock label="Image">
                        <div className="flex flex-col gap-6">
                            <InfoBlock
                                titleWidth={TITLE_WIDTH}
                                title={
                                    <LabelWithInfo
                                        label="Registry To Push Image To"
                                        content="A cluster of several nodes pulls the function's image from it."
                                    />
                                }
                            >
                                <Field>
                                    <PushToRegistryCombobox
                                        projectId={projectId}
                                        env={env}
                                        value={pushToRegistry.value}
                                        onChange={pushToRegistry.onChange}
                                        readOnly={readOnly}
                                        invalid={Boolean(errors.pushToRegistry)}
                                        className={PROJECT_FORM_CONTROL_MAX_WIDTH_CLASS}
                                    />
                                    <FieldError errors={[errors.pushToRegistry]} />
                                    <RegistryCredentialsLink projectId={projectId} />
                                </Field>
                            </InfoBlock>
                        </div>
                    </ContentBlock>

                    {children}
                </fieldset>
            </form>
        </div>
    );
}

type Props = PropsWithChildren<{
    ref?: React.Ref<FunctionSettingsFormRef>;
    projectId: string;
    env: string;
    source: FunctionSource;
    onSubmit: (values: FunctionSettingsFormOutput) => void;
    readOnly?: boolean;
}>;
