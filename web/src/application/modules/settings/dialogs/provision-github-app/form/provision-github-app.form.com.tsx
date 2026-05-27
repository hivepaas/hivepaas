import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import { type FieldErrors, useController, useForm, useWatch } from "react-hook-form";
import { EGithubAppOwnerType } from "~/settings/module-shared/enums";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Checkbox, Field, FieldError, FieldGroup, Input, Tabs, TabsList, TabsTrigger } from "@/components/ui";

import {
    type ProvisionGithubAppFormInput,
    type ProvisionGithubAppFormOutput,
    ProvisionGithubAppFormSchema,
} from "../schemas";

export function ProvisionGithubAppForm({
    isPending,
    loginChecked,
    onLoginCheck,
    onSubmit,
    initialValues,
    showAvailableInProjects,
}: Props) {
    const {
        handleSubmit,
        control,
        formState: { errors },
    } = useForm<ProvisionGithubAppFormInput, unknown, ProvisionGithubAppFormOutput>({
        defaultValues: {
            name: initialValues?.name ?? "",
            ownerType: initialValues?.ownerType ?? EGithubAppOwnerType.Organization,
            org: initialValues?.org ?? "",
            ssoEnabled: initialValues?.ssoEnabled ?? true,
            availableInProjects: initialValues?.availableInProjects ?? true,
            default: initialValues?.default ?? true,
        },
        resolver: zodResolver(ProvisionGithubAppFormSchema),
        mode: "onSubmit",
    });

    const ownerTypeValue = useWatch({ control, name: "ownerType" });
    const showOrganization = ownerTypeValue === EGithubAppOwnerType.Organization;

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({ name: "name", control });
    const { field: ownerType } = useController({ name: "ownerType", control });
    const {
        field: org,
        fieldState: { invalid: isOrgInvalid },
    } = useController({ name: "org", control });
    const { field: ssoEnabled } = useController({ name: "ssoEnabled", control });
    const { field: availableInProjects } = useController({ name: "availableInProjects", control });
    const { field: defaultField } = useController({ name: "default", control });

    function onValid(values: ProvisionGithubAppFormOutput) {
        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<ProvisionGithubAppFormOutput>) {
        console.error(_errors);
    }

    return (
        <form
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onValid, onInvalid)(event);
            }}
            className="flex flex-col gap-6"
        >
            <div className={cn(dashedBorderBox, "flex flex-col gap-5 text-center text-sm leading-6")}>
                <div>
                    <span className="text-orange-500">Important:</span> When you click begin, you will be redirected to
                    the GitHub page where you will be guided on how to set up a GitHub App. Make sure you are logged
                    into GitHub before starting.
                </div>
                <div className="flex justify-start">
                    <Button
                        type="button"
                        onClick={onLoginCheck}
                        disabled={loginChecked}
                    >
                        Github Login Check
                    </Button>
                </div>
            </div>

            <FieldGroup>
                <Field>
                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Name" />}
                    >
                        <Input
                            {...name}
                            aria-invalid={isNameInvalid}
                            placeholder="my github app"
                        />
                        <FieldError errors={[errors.name]} />
                    </InfoBlock>
                </Field>

                <Field>
                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Type" />}
                    >
                        <Tabs
                            value={ownerType.value}
                            onValueChange={value => {
                                ownerType.onChange(value as EGithubAppOwnerType);
                            }}
                        >
                            <TabsList>
                                <TabsTrigger value={EGithubAppOwnerType.Organization}>Organization</TabsTrigger>
                                <TabsTrigger value={EGithubAppOwnerType.User}>User</TabsTrigger>
                            </TabsList>
                        </Tabs>
                    </InfoBlock>
                </Field>

                {showOrganization && (
                    <Field>
                        <InfoBlock
                            titleWidth={220}
                            title={
                                <LabelWithInfo
                                    label="Organization"
                                    isRequired
                                />
                            }
                        >
                            <Input
                                {...org}
                                aria-invalid={isOrgInvalid}
                                placeholder="organization name"
                            />
                            <FieldError errors={[errors.org]} />
                        </InfoBlock>
                    </Field>
                )}

                <Field>
                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="SSO Enabled" />}
                    >
                        <Checkbox
                            checked={ssoEnabled.value}
                            onCheckedChange={checked => {
                                ssoEnabled.onChange(Boolean(checked));
                            }}
                        />
                    </InfoBlock>
                </Field>

                {showAvailableInProjects && (
                    <Field>
                        <InfoBlock
                            titleWidth={220}
                            title={<LabelWithInfo label="Available in Projects" />}
                        >
                            <Checkbox
                                checked={availableInProjects.value}
                                onCheckedChange={checked => {
                                    availableInProjects.onChange(Boolean(checked));
                                }}
                            />
                        </InfoBlock>
                    </Field>
                )}

                <Field>
                    <InfoBlock
                        titleWidth={220}
                        title={<LabelWithInfo label="Default" />}
                    >
                        <Checkbox
                            checked={defaultField.value}
                            onCheckedChange={checked => {
                                defaultField.onChange(Boolean(checked));
                            }}
                        />
                    </InfoBlock>
                </Field>

                <Field>
                    <div className="flex justify-end">
                        <Button
                            type="submit"
                            isLoading={isPending}
                        >
                            Begin Creation Flow
                        </Button>
                    </div>
                </Field>
            </FieldGroup>
        </form>
    );
}

interface Props {
    isPending: boolean;
    loginChecked: boolean;
    onLoginCheck: () => void;
    onSubmit: (values: ProvisionGithubAppFormOutput) => void;
    initialValues?: Partial<ProvisionGithubAppFormInput>;
    showAvailableInProjects: boolean;
}
