import { type PropsWithChildren, useEffect, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import { EJoinNodeMethod } from "~/cluster/module-shared/enums";

import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { ManualMethod, SshMethod } from "../building-blocks";
import { type JoinNewNodeFormInput, type JoinNewNodeFormOutput, JoinNewNodeFormSchema } from "../schemas";

export function JoinNewNodeForm({ onSubmit, onMethodChange, onHasChanges, children }: Props) {
    const [manualCompleted, setManualCompleted] = useState<boolean>(false);
    const methods = useForm<JoinNewNodeFormInput, unknown, JoinNewNodeFormOutput>({
        defaultValues: {
            method: EJoinNodeMethod.RunCommandManually,
            joinAsManager: true,
        },
        resolver: zodResolver(JoinNewNodeFormSchema),
        mode: "onSubmit",
    });

    const {
        handleSubmit,
        control,
        formState: { errors, isDirty },
    } = methods;

    useEffect(() => {
        onHasChanges?.(isDirty);
    }, [isDirty, onHasChanges]);

    const method = useWatch({ control, name: "method" });
    const joinAsManager = useWatch({ control, name: "joinAsManager" });

    const {
        field: methodField,
        fieldState: { invalid: isMethodInvalid },
    } = useController({
        name: "method",
        control,
    });

    const {
        field: joinAsManagerField,
        fieldState: { invalid: isJoinAsManagerInvalid },
    } = useController({
        name: "joinAsManager",
        control,
    });

    function onValid(values: JoinNewNodeFormOutput) {
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<JoinNewNodeFormOutput>) {
        console.log("Invalid", _errors);
        // Optional: log errors or show notification
    }

    return (
        <div className="flex flex-col gap-6">
            <FormProvider {...methods}>
                <form
                    onSubmit={event => {
                        event.preventDefault();
                        void handleSubmit(onValid, onInvalid)(event);
                    }}
                >
                    <FieldGroup>
                        {/* Method Selection */}
                        <Field>
                            <FieldLabel>Please choose the method to join a new node</FieldLabel>
                            <Tabs
                                value={methodField.value}
                                onValueChange={value => {
                                    const newMethod = value as EJoinNodeMethod;
                                    methodField.onChange(newMethod);
                                    onMethodChange?.(newMethod);
                                }}
                            >
                                <TabsList className="w-full">
                                    <TabsTrigger
                                        value={EJoinNodeMethod.RunCommandManually}
                                        className="flex-1"
                                        aria-invalid={isMethodInvalid}
                                        disabled={manualCompleted}
                                    >
                                        Run command manually
                                    </TabsTrigger>
                                    <TabsTrigger
                                        value={EJoinNodeMethod.RunCommandViaSSH}
                                        className="flex-1"
                                        aria-invalid={isMethodInvalid}
                                        disabled={manualCompleted}
                                    >
                                        Run command via SSH
                                    </TabsTrigger>
                                </TabsList>
                            </Tabs>
                            <FieldError errors={[errors.method]} />
                        </Field>

                        {/* Role Selection */}
                        <Field>
                            <FieldLabel>Join node as</FieldLabel>
                            <Tabs
                                value={joinAsManager ? "manager" : "worker"}
                                onValueChange={value => {
                                    joinAsManagerField.onChange(value === "manager");
                                }}
                            >
                                <TabsList className="w-full">
                                    <TabsTrigger
                                        value="manager"
                                        className="flex-1"
                                        aria-invalid={isJoinAsManagerInvalid}
                                        disabled={manualCompleted}
                                    >
                                        Manager
                                    </TabsTrigger>
                                    <TabsTrigger
                                        value="worker"
                                        className="flex-1"
                                        aria-invalid={isJoinAsManagerInvalid}
                                        disabled={manualCompleted}
                                    >
                                        Worker
                                    </TabsTrigger>
                                </TabsList>
                            </Tabs>
                            <FieldError errors={[errors.joinAsManager]} />
                        </Field>

                        {/* Manual Method: Command Display */}
                        {method === EJoinNodeMethod.RunCommandManually && (
                            <ManualMethod
                                onComplete={() => {
                                    setManualCompleted(true);
                                }}
                            />
                        )}

                        {/* SSH Method: SSH Fields */}
                        {method === EJoinNodeMethod.RunCommandViaSSH && <SshMethod />}
                    </FieldGroup>
                    {children}
                </form>
            </FormProvider>
        </div>
    );
}

interface Props extends PropsWithChildren {
    onSubmit: (values: JoinNewNodeFormOutput) => Promise<void> | void;
    onMethodChange?: (method: EJoinNodeMethod) => void;
    onHasChanges?: (dirty: boolean) => void;
}
