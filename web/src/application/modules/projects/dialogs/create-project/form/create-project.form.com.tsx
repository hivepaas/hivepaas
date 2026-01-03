import React, { useState } from "react";

import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, X } from "lucide-react";
import { type FieldErrors, useController, useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";

import { type CreateProjectFormInput, type CreateProjectFormOutput, CreateProjectFormSchema } from "../schemas";

export function CreateProjectForm({ isPending, onSubmit }: Props) {
    const [newTag, setNewTag] = useState("");

    const {
        handleSubmit,
        control,
        formState: { errors },
        watch,
        setValue,
    } = useForm<CreateProjectFormInput, unknown, CreateProjectFormOutput>({
        defaultValues: {
            name: "",
            note: "",
            tags: [],
        },
        resolver: zodResolver(CreateProjectFormSchema),
        mode: "onSubmit",
    });

    const tags = watch("tags");

    const {
        field: name,
        fieldState: { invalid: isNameInvalid },
    } = useController({
        name: "name",
        control,
    });

    const {
        field: note,
        fieldState: { invalid: isNoteInvalid },
    } = useController({
        name: "note",
        control,
    });

    function handleAddTag() {
        if (newTag.trim()) {
            const trimmedTag = newTag.trim();
            if (!tags.includes(trimmedTag)) {
                setValue("tags", [...tags, trimmedTag]);
                setNewTag("");
            }
        }
    }

    function handleRemoveTag(tagToRemove: string) {
        setValue(
            "tags",
            tags.filter(tag => tag !== tagToRemove),
        );
    }

    function onValid(values: CreateProjectFormOutput) {
        void onSubmit(values);
    }

    function onInvalid(_errors: FieldErrors<CreateProjectFormOutput>) {
        // Optional: log errors or show notification
    }

    return (
        <div className="flex flex-col gap-6">
            <form
                onSubmit={event => {
                    event.preventDefault();
                    void handleSubmit(onValid, onInvalid)(event);
                }}
            >
                <FieldGroup>
                    <Field>
                        <FieldLabel htmlFor="name">Name *</FieldLabel>
                        <Input
                            id="name"
                            {...name}
                            placeholder="Enter project name"
                            aria-invalid={isNameInvalid}
                        />
                        <FieldError errors={[errors.name]} />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="note">Note</FieldLabel>
                        <Textarea
                            id="note"
                            {...note}
                            placeholder="Enter project note"
                            rows={4}
                            aria-invalid={isNoteInvalid}
                        />
                        <FieldError errors={[errors.note]} />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="tags">Tags</FieldLabel>
                        <div className="flex flex-col gap-4">
                            <div className="flex gap-2">
                                <Input
                                    placeholder="Enter tag"
                                    value={newTag}
                                    onChange={e => {
                                        setNewTag(e.target.value);
                                    }}
                                    onKeyDown={e => {
                                        if (e.key === "Enter") {
                                            e.preventDefault();
                                            handleAddTag();
                                        }
                                    }}
                                    className="flex-1"
                                />
                                <Button
                                    type="button"
                                    variant="outline"
                                    onClick={handleAddTag}
                                    disabled={!newTag.trim()}
                                >
                                    <Plus className="size-4 mr-2" /> Add
                                </Button>
                            </div>

                            {tags.length > 0 && (
                                <div className="flex flex-wrap gap-2">
                                    {tags.map(tag => (
                                        <div
                                            key={tag}
                                            className="flex items-center gap-1 px-2 py-1 bg-secondary rounded-md text-sm"
                                        >
                                            <span>{tag}</span>
                                            <button
                                                type="button"
                                                onClick={() => {
                                                    handleRemoveTag(tag);
                                                }}
                                                className="hover:bg-destructive/20 rounded p-0.5"
                                            >
                                                <X className="size-3" />
                                            </button>
                                        </div>
                                    ))}
                                </div>
                            )}
                        </div>
                        <FieldError errors={[errors.tags]} />
                    </Field>

                    <Field>
                        <Button
                            type="submit"
                            isLoading={isPending}
                        >
                            Create Project
                        </Button>
                    </Field>
                </FieldGroup>
            </form>
        </div>
    );
}

interface Props {
    isPending: boolean;
    onSubmit: (values: CreateProjectFormOutput) => Promise<void> | void;
}
