import { type PropsWithChildren, useEffect, useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, FormProvider, useController, useForm, useWatch } from "react-hook-form";
import { EJoinNodeMethod } from "~/cluster/module-shared/enums";

import { InfoBlock } from "@application/shared/components";

import { Button } from "@/components/ui/button";
import { DialogActionFooter, DialogBody } from "@/components/ui/dialog";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";

import { ManualMethod, SshMethod } from "../building-blocks";
import { type JoinNewNodeFormInput, type JoinNewNodeFormOutput, JoinNewNodeFormSchema } from "../schemas";

export function JoinNewNodeForm({ readOnly = false, onSubmit, onMethodChange, onHasChanges, children }: Props) {
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
        onHasChanges?.(readOnly ? false : isDirty);
    }, [isDirty, onHasChanges, readOnly]);

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
        if (readOnly) {
            return;
        }

        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<JoinNewNodeFormOutput>) {
        console.log("Invalid", _errors);
        // Optional: log errors or show notification
    }

    return (
        <FormProvider {...methods}>
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
                className="min-h-0 flex flex-1 flex-col"
            >
                <DialogBody>
                    <FieldGroup>
                        {/* Method Selection */}
                        <Field>
                            <FieldLabel>Choose the method to join a new node</FieldLabel>
                            <Tabs
                                value={methodField.value}
                                onValueChange={value => {
                                    if (readOnly) {
                                        return;
                                    }

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
                                        disabled={readOnly || manualCompleted}
                                    >
                                        Run command manually
                                    </TabsTrigger>
                                    <TabsTrigger
                                        value={EJoinNodeMethod.RunCommandViaSSH}
                                        className="flex-1"
                                        aria-invalid={isMethodInvalid}
                                        disabled={readOnly || manualCompleted}
                                    >
                                        Run command via SSH
                                    </TabsTrigger>
                                </TabsList>
                            </Tabs>
                            <FieldError errors={[errors.method]} />
                        </Field>

                        {/* Role Selection */}
                        <Field>
                            <InfoBlock
                                title="Join node as"
                                titleWidth={150}
                            >
                                <Tabs
                                    value={joinAsManager ? "manager" : "worker"}
                                    onValueChange={value => {
                                        if (readOnly) {
                                            return;
                                        }

                                        joinAsManagerField.onChange(value === "manager");
                                    }}
                                >
                                    <TabsList>
                                        <TabsTrigger
                                            value="manager"
                                            className="flex-1"
                                            aria-invalid={isJoinAsManagerInvalid}
                                            disabled={readOnly || manualCompleted}
                                        >
                                            Manager
                                        </TabsTrigger>
                                        <TabsTrigger
                                            value="worker"
                                            className="flex-1"
                                            aria-invalid={isJoinAsManagerInvalid}
                                            disabled={readOnly || manualCompleted}
                                        >
                                            Worker
                                        </TabsTrigger>
                                    </TabsList>
                                </Tabs>
                                <FieldError errors={[errors.joinAsManager]} />
                            </InfoBlock>
                        </Field>

                        {/* Manual Method: Command Display */}
                        {method === EJoinNodeMethod.RunCommandManually && (
                            <ManualMethod
                                readOnly={readOnly}
                                onComplete={() => {
                                    setManualCompleted(true);
                                }}
                            />
                        )}

                        {/* SSH Method: SSH Fields */}
                        {method === EJoinNodeMethod.RunCommandViaSSH && <SshMethod readOnly={readOnly} />}
                    </FieldGroup>
                    {children}
                </DialogBody>
                {method === EJoinNodeMethod.RunCommandViaSSH ? (
                    <DialogActionFooter>
                        <Button
                            type="submit"
                            disabled={readOnly}
                        >
                            Join Node
                        </Button>
                    </DialogActionFooter>
                ) : null}
            </form>
        </FormProvider>
    );
}

interface Props extends PropsWithChildren {
    readOnly?: boolean;
    onSubmit: (values: JoinNewNodeFormOutput) => Promise<void> | void;
    onMethodChange?: (method: EJoinNodeMethod) => void;
    onHasChanges?: (dirty: boolean) => void;
}
