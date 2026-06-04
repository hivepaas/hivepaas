import { useEffect } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { dashedBorderBox } from "@lib/styles";
import { cn } from "@lib/utils";
import {
    Controller,
    type FieldErrors,
    type FieldError as HookFormFieldError,
    type UseFormRegisterReturn,
    useForm,
} from "react-hook-form";

import { SupportFeedbackCategory } from "@application/shared/enums";

import { Button, Field, FieldError, Input } from "@/components/ui";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";

import {
    CreateFeedbackFormSchema,
    type CreateFeedbackFormSchemaInput,
    type CreateFeedbackFormSchemaOutput,
    emptyCreateFeedbackFormDefaults,
} from "../schemas";

const CATEGORY_OPTIONS = [
    { label: "General", value: SupportFeedbackCategory.General },
    { label: "Security Report", value: SupportFeedbackCategory.SecurityReport },
    { label: "Bug/Issue Report", value: SupportFeedbackCategory.BugIssueReport },
    { label: "Licensing", value: SupportFeedbackCategory.Licensing },
] as const;

export function CreateFeedbackForm({ defaultValues, isPending = false, onSubmit }: Props) {
    const {
        control,
        formState: { errors },
        handleSubmit,
        register,
        reset,
    } = useForm<CreateFeedbackFormSchemaInput, unknown, CreateFeedbackFormSchemaOutput>({
        defaultValues: {
            ...emptyCreateFeedbackFormDefaults,
            ...defaultValues,
        },
        resolver: zodResolver(CreateFeedbackFormSchema),
        mode: "onSubmit",
    });

    useEffect(() => {
        reset({
            ...emptyCreateFeedbackFormDefaults,
            ...defaultValues,
        });
    }, [defaultValues, reset]);

    function onInvalid(_errors: FieldErrors<CreateFeedbackFormSchemaInput>) {
        return undefined;
    }

    return (
        <form
            className="flex flex-col gap-5"
            onSubmit={event => {
                event.preventDefault();
                void handleSubmit(onSubmit, onInvalid)(event);
            }}
        >
            <div className={cn(dashedBorderBox, "text-center text-sm leading-6 px-4 py-3")}>
                <span className="text-orange-500">Note:</span> If you want to interact with the LocalPaaS development
                team and user community, you can join our chat channel at{" "}
                <span className="text-primary">this link</span>.
            </div>

            <Field className="grid grid-cols-[190px_minmax(0,1fr)] items-start gap-4">
                <label
                    htmlFor="feedback-category"
                    className="pt-2 font-medium"
                >
                    Category
                </label>
                <div>
                    <Controller
                        control={control}
                        name="category"
                        render={({ field }) => (
                            <Select
                                value={field.value}
                                onValueChange={field.onChange}
                            >
                                <SelectTrigger
                                    id="feedback-category"
                                    aria-invalid={Boolean(errors.category)}
                                >
                                    <SelectValue placeholder="select category" />
                                </SelectTrigger>
                                <SelectContent>
                                    {CATEGORY_OPTIONS.map(option => (
                                        <SelectItem
                                            key={option.value}
                                            value={option.value}
                                        >
                                            {option.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        )}
                    />
                    <FieldError errors={[errors.category]} />
                </div>
            </Field>

            <FeedbackInput
                error={errors.name}
                label="Your Name"
                placeholder="your name"
                registration={register("name")}
            />
            <FeedbackInput
                error={errors.email}
                label="Your Email"
                placeholder="optional"
                registration={register("email")}
            />
            <FeedbackInput
                error={errors.company}
                label="Your Company"
                placeholder="optional"
                registration={register("company")}
            />
            <FeedbackInput
                required
                error={errors.subject}
                label="Subject"
                placeholder="subject"
                registration={register("subject")}
            />

            <Field className="grid grid-cols-[190px_minmax(0,1fr)] items-start gap-4">
                <label
                    htmlFor="feedback-description"
                    className="pt-2 font-medium"
                >
                    Description <span className="text-destructive">*</span>
                </label>
                <div>
                    <Textarea
                        id="feedback-description"
                        aria-invalid={Boolean(errors.description)}
                        className="min-h-[220px]"
                        minRows={3}
                        maxRows={6}
                        {...register("description")}
                    />
                    <FieldError errors={[errors.description]} />
                </div>
            </Field>

            <div className="flex justify-end pt-2">
                <Button
                    type="submit"
                    disabled={isPending}
                    className="min-w-[130px]"
                >
                    Submit
                </Button>
            </div>
        </form>
    );
}

function FeedbackInput({ error, label, placeholder, registration, required = false }: FeedbackInputProps) {
    return (
        <Field className="grid grid-cols-[190px_minmax(0,1fr)] items-start gap-4">
            <label
                htmlFor={`feedback-${registration.name}`}
                className="pt-2 font-medium"
            >
                {label} {required ? <span className="text-destructive">*</span> : null}
            </label>
            <div>
                <Input
                    id={`feedback-${registration.name}`}
                    aria-invalid={Boolean(error)}
                    placeholder={placeholder}
                    {...registration}
                />
                <FieldError errors={[error]} />
            </div>
        </Field>
    );
}

interface Props {
    defaultValues?: Partial<CreateFeedbackFormSchemaInput>;
    isPending?: boolean;
    onSubmit: (values: CreateFeedbackFormSchemaOutput) => void;
}

type FeedbackInputProps = {
    error?: HookFormFieldError;
    label: string;
    placeholder: string;
    registration: UseFormRegisterReturn;
    required?: boolean;
};
