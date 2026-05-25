import { useEffect } from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { z } from "zod";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Field, FieldError, FieldGroup, Input } from "@/components/ui";

const DEFAULT_TEST_SEND_MAIL_VALUES = {
    testRecipient: "",
    testSubject: "test message from LocalPaaS",
    testContent: "this is test email content",
};

const TestSendMailFormSchema = z.object({
    testRecipient: z.string().trim().min(1, "Recipient is required").email("Recipient must be a valid email"),
    testSubject: z.string().trim().min(1, "Subject is required"),
    testContent: z.string().trim().min(1, "Content is required"),
});

type TestSendMailFormInput = z.input<typeof TestSendMailFormSchema>;
export type TestSendMailFormOutput = z.output<typeof TestSendMailFormSchema>;

export function TestSendMailDialog({ open, isPending, testStatus, onOpenChange, onSubmit }: Props) {
    const {
        handleSubmit,
        control,
        reset,
        formState: { errors },
    } = useForm<TestSendMailFormInput, unknown, TestSendMailFormOutput>({
        defaultValues: DEFAULT_TEST_SEND_MAIL_VALUES,
        resolver: zodResolver(TestSendMailFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        if (open) {
            reset(DEFAULT_TEST_SEND_MAIL_VALUES);
        }
    }, [open, reset]);

    const {
        field: testRecipient,
        fieldState: { invalid: isTestRecipientInvalid },
    } = useController({ name: "testRecipient", control });
    const {
        field: testSubject,
        fieldState: { invalid: isTestSubjectInvalid },
    } = useController({ name: "testSubject", control });
    const {
        field: testContent,
        fieldState: { invalid: isTestContentInvalid },
    } = useController({ name: "testContent", control });

    function onValid(values: TestSendMailFormOutput) {
        onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<TestSendMailFormOutput>) {
        console.error(_errors);
    }

    function handleOpenChange(nextOpen: boolean) {
        if (isPending) {
            return;
        }

        onOpenChange(nextOpen);
    }

    return (
        <Dialog
            open={open}
            onOpenChange={handleOpenChange}
        >
            <DialogContent className="min-w-[390px] w-[680px] max-h-[90vh] overflow-y-auto">
                <DialogHeader>
                    <DialogTitle>Test Send Mail</DialogTitle>
                </DialogHeader>

                <form
                    onSubmit={event => {
                        event.preventDefault();
                        void handleSubmit(onValid, onInvalid)(event);
                    }}
                >
                    <FieldGroup>
                        <InfoBlock
                            titleWidth={160}
                            title={
                                <LabelWithInfo
                                    label="Recipient"
                                    isRequired
                                />
                            }
                        >
                            <Field>
                                <Input
                                    {...testRecipient}
                                    placeholder="my-email@domain.name"
                                    aria-invalid={isTestRecipientInvalid}
                                />
                                <FieldError errors={[errors.testRecipient]} />
                            </Field>
                        </InfoBlock>

                        <InfoBlock
                            titleWidth={160}
                            title={
                                <LabelWithInfo
                                    label="Subject"
                                    isRequired
                                />
                            }
                        >
                            <Field>
                                <Input
                                    {...testSubject}
                                    aria-invalid={isTestSubjectInvalid}
                                />
                                <FieldError errors={[errors.testSubject]} />
                            </Field>
                        </InfoBlock>

                        <InfoBlock
                            titleWidth={160}
                            title={
                                <LabelWithInfo
                                    label="Content"
                                    isRequired
                                />
                            }
                        >
                            <Field>
                                <Textarea
                                    {...testContent}
                                    minRows={6}
                                    aria-invalid={isTestContentInvalid}
                                />
                                <FieldError errors={[errors.testContent]} />
                            </Field>
                        </InfoBlock>

                        <Field>
                            <div className="flex items-center gap-4">
                                <Button
                                    type="submit"
                                    isLoading={isPending}
                                >
                                    Send Test Email
                                </Button>
                                {testStatus === "succeeded" && (
                                    <span className="text-sm font-medium text-green-600">Succeeded</span>
                                )}
                                {testStatus === "failed" && (
                                    <span className="text-sm font-medium text-destructive">Failed</span>
                                )}
                            </div>
                        </Field>
                    </FieldGroup>
                </form>
            </DialogContent>
        </Dialog>
    );
}

interface Props {
    open: boolean;
    isPending: boolean;
    testStatus: "idle" | "succeeded" | "failed";
    onOpenChange: (open: boolean) => void;
    onSubmit: (values: TestSendMailFormOutput) => void;
}
