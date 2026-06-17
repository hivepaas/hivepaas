import { zodResolver } from "@hookform/resolvers/zod";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { InfoBlock, LabelWithInfo } from "@application/shared/components";

import {
    Button,
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
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
} from "../../schemas";

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
            <DialogFixedContent className="sm:max-w-[520px]">
                <DialogHeader>
                    <DialogTitle>Test DNS Access</DialogTitle>
                </DialogHeader>
                <form
                    onSubmit={event => {
                        event.preventDefault();
                        void handleSubmit(onValid, onInvalid)(event);
                    }}
                >
                    <DialogBody>
                        <InfoBlock
                            titleWidth={160}
                            title={<LabelWithInfo label="Your Domain" />}
                        >
                            <FieldGroup>
                                <Field>
                                    <Input
                                        {...testDomain}
                                        placeholder="mydomain.com"
                                        aria-invalid={invalid}
                                    />
                                    <FieldError errors={[errors.testDomain]} />
                                </Field>
                            </FieldGroup>
                        </InfoBlock>
                    </DialogBody>
                    <DialogActionFooter>
                        <div className="flex justify-end">
                            <Button
                                type="submit"
                                isLoading={isPending}
                                className="min-w-[100px]"
                            >
                                Test
                            </Button>
                        </div>
                    </DialogActionFooter>
                </form>
            </DialogFixedContent>
        </Dialog>
    );
}

interface Props {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    isPending: boolean;
    onSubmit: (domain: string) => void;
}
