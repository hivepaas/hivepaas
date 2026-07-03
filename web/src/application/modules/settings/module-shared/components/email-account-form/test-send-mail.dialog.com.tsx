import { useEffect } from "react";

import * as DialogPrimitive from "@radix-ui/react-dialog";
import { Dialog, DialogHeader, DialogOverlay, DialogPortal, DialogTitle } from "@components/ui/dialog";
import { Textarea } from "@components/ui/textarea";
import { zodResolver } from "@hookform/resolvers/zod";
import { cn } from "@lib/utils";
import { XIcon } from "lucide-react";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { z } from "zod";
import { SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import { Button, Field, FieldError, FieldGroup, Input } from "@/components/ui";

const DEFAULT_TEST_SEND_MAIL_VALUES = {
    testRecipient: "",
    testSubject: "test message from HivePaaS",
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
            <DialogPortal>
                <DialogOverlay />
                <DialogPrimitive.Content
                    className={cn(
                        "bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed top-[50%] left-[50%] z-50 grid max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] rounded-lg border shadow-lg duration-200",
                        "max-h-[90vh] min-w-[390px] w-[680px] overflow-hidden flex flex-col gap-0 p-0",
                        "[&>[data-slot=dialog-header]]:shrink-0 [&>[data-slot=dialog-header]]:px-6 [&>[data-slot=dialog-header]]:pt-6 [&>[data-slot=dialog-header]]:pb-4",
                    )}
                >
                    <DialogHeader>
                        <DialogTitle>Test Send Mail</DialogTitle>
                    </DialogHeader>

                    <form
                        onSubmit={event => {
                            event.preventDefault();
                            void handleSubmit(onValid, onInvalid)(event);
                        }}
                        className="min-h-0 flex flex-1 flex-col"
                    >
                        <div className="min-h-0 flex-1 overflow-y-auto px-6 py-4">
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
                                    <Field className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}>
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
                                    <Field className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}>
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
                                    <Field className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                        <Textarea
                                            {...testContent}
                                            minRows={6}
                                            aria-invalid={isTestContentInvalid}
                                        />
                                        <FieldError errors={[errors.testContent]} />
                                    </Field>
                                </InfoBlock>
                            </FieldGroup>
                        </div>
                        <div className="shrink-0 px-0 mt-6 pb-6 pr-6 flex justify-end items-center gap-3">
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
                    </form>
                    <DialogPrimitive.Close className="ring-offset-background focus:ring-ring data-[state=open]:bg-accent data-[state=open]:text-muted-foreground absolute top-4 right-4 rounded-xs opacity-70 transition-opacity hover:opacity-100 focus:ring-2 focus:ring-offset-2 focus:outline-hidden disabled:pointer-events-none [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4">
                        <XIcon />
                        <span className="sr-only">Close</span>
                    </DialogPrimitive.Close>
                </DialogPrimitive.Content>
            </DialogPortal>
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
