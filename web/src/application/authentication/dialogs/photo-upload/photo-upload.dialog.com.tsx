import { useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";

import classNames from "classnames/bind";

import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";
import { Slider } from "@components/ui/slider";
import "cropperjs/dist/cropper.css";
import { Pencil, RotateCcw, Trash2, ZoomIn, ZoomOut } from "lucide-react";
import Cropper, { type ReactCropperElement } from "react-cropper";

import styles from "./photo-upload.dialog.module.scss";

const cx = classNames.bind(styles);
interface CroppedImageResult {
    fileName: string;
    dataBase64: string;
}

interface PhotoUploadDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onConfirm: (result: CroppedImageResult) => void;
    initialImage?: string | null;
}

export function PhotoUploadDialog({ open, onOpenChange, onConfirm, initialImage = null }: PhotoUploadDialogProps) {
    const cropperRef = useRef<ReactCropperElement | null>(null);
    const fileInputRef = useRef<HTMLInputElement | null>(null);
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [imageSrc, setImageSrc] = useState<string | null>(initialImage);
    const [isCropping, setIsCropping] = useState(false);
    const [rotation, setRotation] = useState(0); // degrees 0..180..360
    const [zoom, setZoom] = useState(1); // 0.1 .. 3
    const [loaded, setLoaded] = useState(false);

    const fileName = useMemo(() => {
        return selectedFile?.name ?? "image.png";
    }, [selectedFile]);

    useEffect(() => {
        if (open) {
            setImageSrc(initialImage ?? null);
            setRotation(0);
            setZoom(1);
            setSelectedFile(null);
            setLoaded(false);
        }
    }, [open, initialImage]);

    useLayoutEffect(() => {
        if (!loaded || !cropperRef.current) {
            return;
        }

        const { cropper } = cropperRef.current;

        const container: Partial<Cropper.ContainerData> = cropper.getContainerData();
        const cropBox: Partial<Cropper.CropBoxData> = cropper.getCropBoxData();
        const image: Partial<Cropper.ImageData> = cropper.getImageData();

        if (
            container.width === undefined ||
            container.height === undefined ||
            cropBox.width === undefined ||
            cropBox.height === undefined ||
            image.width === undefined ||
            image.height === undefined ||
            image.naturalWidth === undefined ||
            image.naturalHeight === undefined
        ) {
            return;
        }

        /**
         * Calculate zoom ratio, where 0 is 100% of the container size.
         */
        const ratio =
            (1 + zoom / 50) *
            Math.min(
                container.width / image.naturalWidth,
                container.height / image.naturalHeight, //
            );

        /**
         * Zoom from 100% to 300%. Center the zoom on the container.
         */
        cropper.zoomTo(ratio, {
            x: container.width / 2,
            y: container.height / 2,
        });

        /**
         * Limit the crop box to the image size.
         */
        if (cropBox.width > image.width || cropBox.height > image.height) {
            cropper.setCropBoxData({
                width: Math.min(cropBox.width, image.width),
                height: Math.min(cropBox.height, image.height),
            });
        }
    }, [loaded, zoom]);

    function onFileChange(e: React.ChangeEvent<HTMLInputElement>) {
        const file = e.target.files?.[0];

        if (file === undefined) {
            return;
        }

        const reader = new FileReader();
        reader.onload = () => {
            setSelectedFile(file);
            setImageSrc(typeof reader.result === "string" ? reader.result : null);
            setRotation(0);
            setZoom(1);
            if (cropperRef.current) {
                cropperRef.current.cropper.reset();
            }
        };
        reader.readAsDataURL(file);
    }

    function triggerSelectFile() {
        fileInputRef.current?.click();
    }

    function handleDelete() {
        setSelectedFile(null);
        setImageSrc(null);
        setRotation(0);
        setZoom(1);
        if (fileInputRef.current) {
            fileInputRef.current.value = "";
        }
        if (cropperRef.current) {
            cropperRef.current.cropper.clear();
            cropperRef.current.cropper.reset();
        }
    }

    function handleRotateTo(deg: number) {
        setRotation(deg);
        if (cropperRef.current) {
            cropperRef.current.cropper.rotateTo(deg);
        }
    }

    function handleZoomTo(value: number) {
        setZoom(value);
        if (cropperRef.current) {
            cropperRef.current.cropper.zoomTo(value);
        }
    }

    function handleReset() {
        setRotation(0);
        setZoom(1);
        if (cropperRef.current) {
            cropperRef.current.cropper.reset();
        }
    }

    function handleConfirm() {
        if (cropperRef.current == null) {
            return;
        }

        try {
            setIsCropping(true);
            const canvas = cropperRef.current.cropper.getCroppedCanvas({
                imageSmoothingEnabled: true,
                imageSmoothingQuality: "high",
            });
            const dataBase64 = canvas.toDataURL("image/png", 0.92);

            onConfirm({
                fileName,
                dataBase64,
            });
            onOpenChange(false);
        } finally {
            setIsCropping(false);
        }
    }

    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
            modal
        >
            <DialogContent className="sm:max-w-2xl">
                <DialogHeader>
                    <DialogTitle>Profile Photo</DialogTitle>
                    <DialogDescription>Adjust the crop area. Use tools to rotate/zoom.</DialogDescription>
                </DialogHeader>

                {/* Hidden file input to be triggered by toolbar button */}
                <input
                    ref={fileInputRef}
                    type="file"
                    accept="image/*"
                    onChange={onFileChange}
                    className="hidden"
                />

                <div className="relative flex flex-col gap-3">
                    {/* Top-left floating toolbar */}
                    <div className="pointer-events-none absolute left-3 top-3 z-10 flex gap-3">
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            className="pointer-events-auto"
                            onClick={triggerSelectFile}
                        >
                            <Pencil />
                            Choose Photo
                        </Button>
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            className="pointer-events-auto"
                            onClick={handleDelete}
                            disabled={!imageSrc}
                        >
                            <Trash2 />
                            Delete
                        </Button>
                    </div>

                    {/* Cropper area with circular mask - square container */}
                    <div className="relative w-full">
                        {imageSrc ? (
                            <Cropper
                                ref={cropperRef}
                                style={{
                                    width: "100%",
                                    height: "400px",
                                }}
                                className={cx("cropper")}
                                src={imageSrc}
                                aspectRatio={1}
                                checkOrientation={false}
                                minCropBoxHeight={100}
                                minCropBoxWidth={100}
                                guides={false}
                                center={false}
                                viewMode={2}
                                zoomOnTouch={false}
                                zoomOnWheel={false}
                                dragMode="move"
                                background={false}
                                toggleDragModeOnDblclick={false}
                                onLoad={() => {
                                    setLoaded(true);
                                }}
                            />
                        ) : (
                            <div className="flex w-full h-[400px] items-center justify-center rounded-md border text-muted-foreground">
                                Select an image to start
                            </div>
                        )}

                        {/* Vertical rotation slider on the right */}
                        <div className="absolute right-2 top-1/2 z-10 -translate-y-1/2">
                            <div className="flex h-[320px] flex-col items-center justify-between">
                                <span className="text-xs text-muted-foreground">180°</span>
                                <Slider
                                    orientation="vertical"
                                    min={0}
                                    max={360}
                                    step={1}
                                    value={[rotation]}
                                    onValueChange={vals => {
                                        handleRotateTo(vals[0] ?? 0);
                                    }}
                                    className="h-[300px]"
                                />
                                <span className="text-xs text-muted-foreground">0°</span>
                            </div>
                        </div>
                    </div>

                    {/* Bottom toolbar: zoom slider + controls */}
                    <div className="flex items-center gap-3 px-2">
                        <Button
                            type="button"
                            variant="outline"
                            size="icon-sm"
                            onClick={() => {
                                handleZoomTo(Math.max(0.1, zoom - 0.1));
                            }}
                            disabled={!imageSrc}
                        >
                            <ZoomOut />
                        </Button>
                        <Slider
                            min={0.1}
                            max={3}
                            step={0.01}
                            value={[zoom]}
                            onValueChange={vals => {
                                handleZoomTo(vals[0] ?? 1);
                            }}
                            className="w-full"
                        />
                        <Button
                            type="button"
                            variant="outline"
                            size="icon-sm"
                            onClick={() => {
                                handleZoomTo(Math.min(3, zoom + 0.1));
                            }}
                            disabled={!imageSrc}
                        >
                            <ZoomIn />
                        </Button>
                        <div className="grow" />
                        <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={handleReset}
                            disabled={!imageSrc}
                        >
                            <RotateCcw />
                            Reset
                        </Button>
                    </div>
                </div>

                <DialogFooter>
                    <Button
                        variant="outline"
                        onClick={() => {
                            onOpenChange(false);
                        }}
                    >
                        Cancel
                    </Button>
                    <Button
                        disabled={!imageSrc || isCropping}
                        isLoading={isCropping}
                        onClick={() => {
                            handleConfirm();
                        }}
                    >
                        Save
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
