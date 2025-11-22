import React, { useRef, useState } from "react";

import classnames from "classnames/bind";

import { Button } from "@components/ui";
import "cropperjs/dist/cropper.css";
import { type LucideIcon, Trash2, UploadIcon } from "lucide-react";
import { type ReactCropperElement } from "react-cropper";
import { toast } from "sonner";

import { Editor, PhotoPlaceholder } from "../building-blocks";

import styles from "./photo-editor.module.scss";

const cx = classnames.bind(styles);

const SUPPORTED_FORMATS = ["image/jpg", "image/jpeg", "image/png", "image/webp", "image/gif"];

const DEFAULT_MAX_SIZE = 10 * 1024 * 1024; // 5MB

export function PhotoEditor({
    ref,
    src,
    maxSize = DEFAULT_MAX_SIZE,
    hideIcon,
    placeholderIcon,
    onEditingStart,
    onEditingEnd,
    onDelete,
}: Props) {
    const [image, setImage] = useState<string | null>(src);
    const fileInputRef = useRef<HTMLInputElement>(null);

    function handleFileChoose(file: File) {
        const reader = new FileReader();

        reader.onload = () => {
            const { result } = reader;

            if (typeof result === "string") {
                setImage(result);
            }
        };

        reader.onerror = () => {
            toast.error("Failed to load the image");

            setImage(null);
        };

        reader.readAsDataURL(file);
    }

    function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
        const file = e.target.files?.[0];

        if (file === undefined) {
            return;
        }

        if (file.size > maxSize) {
            toast.error(`File size should not exceed ${maxSize / 1024 / 1024}MB`);
            return;
        }

        if (!SUPPORTED_FORMATS.includes(file.type)) {
            toast.error(`File format should be one of the following: ${SUPPORTED_FORMATS.join(", ")}`);
            return;
        }

        handleFileChoose(file);
    }

    function triggerSelectFile() {
        fileInputRef.current?.click();
    }

    return (
        <div className={cx("photo-editor")}>
            {image ? (
                <Editor
                    ref={ref}
                    src={image}
                    onEditingStart={onEditingStart}
                    onEditingEnd={onEditingEnd}
                    onError={() => {
                        toast.error("Failed to load the image");

                        setImage(null);
                    }}
                />
            ) : (
                <PhotoPlaceholder
                    hideIcon={hideIcon}
                    icon={placeholderIcon}
                />
            )}

            <div className={cx("photo-actions")}>
                <input
                    ref={fileInputRef}
                    type="file"
                    accept=".png,.jpg,.jpeg,.webp,.gif"
                    onChange={handleFileChange}
                    className="hidden"
                />
                <Button
                    size="sm"
                    variant="outline"
                    onClick={e => {
                        e.preventDefault();
                        triggerSelectFile();
                    }}
                >
                    <UploadIcon className={cx("upload-icon")} />
                    Choose Photo
                </Button>

                <Button
                    size="sm"
                    variant="destructive"
                    disabled={!image}
                    onClick={() => {
                        if (onDelete) {
                            onDelete();
                        } else {
                            setImage(null);
                        }
                    }}
                >
                    <Trash2 className={cx("trash-icon")} />
                    Delete
                </Button>
            </div>
        </div>
    );
}

export type PhotoEditorRef = ReactCropperElement;

interface Props {
    ref?: React.Ref<ReactCropperElement | HTMLImageElement>;
    src: string | null;
    maxSize?: number;
    onEditingStart?: () => void;
    onEditingEnd?: () => void;
    hideIcon?: boolean;
    placeholderIcon?: LucideIcon;
    onDelete?: () => void;
}
