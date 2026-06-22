import * as DialogPrimitive from "@radix-ui/react-dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { cn } from "@lib/utils";
import { XIcon } from "lucide-react";
import { type FieldErrors, useController, useForm } from "react-hook-form";
import { SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS } from "~/settings/module-shared/constants/settings-form-layout.constants";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import {
    Button,
    Dialog,
    DialogHeader,
    DialogOverlay,
    DialogPortal,
    DialogTitle,
    Field,
    FieldError,
    FieldGroup,
    Input,
} from "@/components/ui";

import {
    type TestAcmeDnsProviderAccessFormInput,
    type TestAcmeDnsProviderAccessFormOutput,
    TestAcmeDnsProviderAccessFormSchema,
} from "../create-or-edit-acme-dns-provider.form.schema";

export function TestAcmeDnsProviderAccessDialog({ open, onOpenChange, isPending, onSubmit }: Props) {
    const {
        control,
        handleSubmit,
        formState: { errors },
    } = useForm<TestAcmeDnsProviderAccessFormInput, unknown, TestAcmeDnsProviderAccessFormOutput>({
        defaultValues: {
            testDomain: "",
        },
        resolver: zodResolver(TestAcmeDnsProviderAccessFormSchema),
        mode: "onSubmit",
    });

    const {
        field: testDomain,
        fieldState: { invalid },
    } = useController({ name: "testDomain", control });

    function onValid(values: TestAcmeDnsProviderAccessFormOutput) {
        onSubmit(values.testDomain);
    }

    function onInvalid(_errors: FieldErrors<TestAcmeDnsProviderAccessFormOutput>) {
        console.error(_errors);
    }

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogPortal>
                <DialogOverlay />
                <DialogPrimitive.Content
                    className={cn(
                        "bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] rounded-lg border shadow-lg duration-200 sm:max-w-[520px]",
                        "max-h-[90vh] overflow-hidden flex flex-col gap-0 p-0",
                        "[&>[data-slot=dialog-header]]:shrink-0 [&>[data-slot=dialog-header]]:px-6 [&>[data-slot=dialog-header]]:pt-6 [&>[data-slot=dialog-header]]:pb-4",
                    )}
                >
                    <DialogHeader>
                        <DialogTitle>Test DNS Access</DialogTitle>
                    </DialogHeader>
                    <form
                        onSubmit={event => {
                            event.preventDefault();
                            void handleSubmit(onValid, onInvalid)(event);
                        }}
                    >
                        <div className="min-h-0 flex-1 overflow-y-auto px-6 py-4">
                            <InfoBlock
                                titleWidth={160}
                                title={<LabelWithInfo label="Your Domain" />}
                            >
                                <FieldGroup>
                                    <Field className={SETTINGS_FORM_CONTROL_MAX_WIDTH_CLASS}>
                                        <Input
                                            {...testDomain}
                                            placeholder="mydomain.com"
                                            aria-invalid={invalid}
                                        />
                                        <FieldError errors={[errors.testDomain]} />
                                    </Field>
                                </FieldGroup>
                            </InfoBlock>
                        </div>
                        <div className="shrink-0 px-0 mt-6 pb-6 flex justify-end">
                            <div className="flex justify-end">
                                <Button
                                    type="submit"
                                    isLoading={isPending}
                                    className="min-w-[100px]"
                                >
                                    Test
                                </Button>
                            </div>
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
    onOpenChange: (open: boolean) => void;
    isPending: boolean;
    onSubmit: (domain: string) => void;
}
