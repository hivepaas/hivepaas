import React, { type PropsWithChildren, useRef } from "react";

import classnames from "classnames/bind";

import { zodResolver } from "@hookform/resolvers/zod";
import { ImageService } from "@infrastructure/services";
import { Ok, type Result, match } from "oxide.ts";
import { type FieldErrors, useForm } from "react-hook-form";
import { toast } from "sonner";

import { PhotoEditor, type PhotoEditorRef } from "@application/shared/components/photo-editor/component";

import {
    type UploadPhotoFormInput,
    type UploadPhotoFormOutput,
    UploadPhotoFormSchema,
} from "../schemas/upload-photo.form.schema";

import styles from "./upload-photo.form.module.scss";

const cx = classnames.bind(styles);

export function UploadPhotoForm({ photo, onSubmit, onEditingStart, onEditingEnd, children }: Props) {
    const { handleSubmit: originalHandleSubmit, setValue } = useForm<
        UploadPhotoFormInput,
        unknown,
        UploadPhotoFormOutput
    >({
        defaultValues: {
            photo: null,
        },
        resolver: zodResolver(UploadPhotoFormSchema),
        mode: "onSubmit",
    });

    const editorRef = useRef<PhotoEditorRef>(null);

    async function getImageFile(): Promise<Result<{ file: File | null }, Error>> {
        const editor = editorRef.current?.cropper;

        if (!editor) {
            return Ok({ file: null });
        }

        const canvas = editor.getCroppedCanvas({
            width: 512,
            height: 512,
        });

        const bitmap = await createImageBitmap(canvas);

        return ImageService.convert(bitmap, {
            filename: "profile-photo",
            type: "image/webp",
            quality: 0.9,
        });
    }

    async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
        const result = await getImageFile();
        match(result, {
            Ok: ({ file }) => {
                setValue("photo", file);
                void originalHandleSubmit(onValid, onInvalid)(event);
            },
            Err: error => {
                toast.error(error.message);
            },
        });
    }

    function onValid(values: UploadPhotoFormOutput) {
        onSubmit(values);
    }

    function onInvalid(errors: FieldErrors<UploadPhotoFormInput>) {
        console.error(errors);
    }

    return (
        <form
            className={cx("upload-photo-form")}
            onSubmit={event => {
                event.preventDefault();

                void handleSubmit(event);
            }}
        >
            <div className={cx("photo")}>
                <PhotoEditor
                    ref={editorRef}
                    src={photo}
                    onEditingStart={onEditingStart}
                    onEditingEnd={onEditingEnd}
                />
            </div>

            {children}
        </form>
    );
}

type Props = PropsWithChildren<{
    photo: string | null;
    onSubmit: (values: UploadPhotoFormOutput) => void;
    onEditingStart?: () => void;
    onEditingEnd?: () => void;
}>;
