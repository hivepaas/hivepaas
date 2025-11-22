import React, { useImperativeHandle, useLayoutEffect, useRef, useState } from "react";

import classnames from "classnames/bind";

import { Button } from "@components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@components/ui/dialog";
import { Slider } from "@components/ui/slider";
import "cropperjs/dist/cropper.css";
import { Eye, RotateCcw, ZoomIn, ZoomOut } from "lucide-react";
import { Cropper, type ReactCropperElement } from "react-cropper";

import styles from "./editor.module.scss";

const cx = classnames.bind(styles);

const zoomFormatter = (value: number) => {
    if (!value) {
        return "100%";
    }

    return `${100 + value * 2}%`;
};

const rotateFormatter = (value: number) => {
    if (!value) {
        return "0°";
    }

    return `${-value}°`;
};

/**
 * For development purposes only.
 */
const SHOW_PREVIEW: boolean = false;

type PreviewState =
    | {
          open: true;
          src: string;
      }
    | {
          open: false;
          src: undefined;
      };

function View({ ref, src, onError, onEditingStart, onEditingEnd }: Props) {
    const [loaded, setLoaded] = useState(false);
    const [zoom, setZoom] = useState(0);
    const [rotate, setRotate] = useState(0);

    const [previewOpen, setPreviewOpen] = useState<PreviewState>({
        open: false,
        src: undefined,
    });

    const editorRef = useRef<ReactCropperElement>(null);

    useLayoutEffect(() => {
        setLoaded(false);
        setZoom(0);
        setRotate(0);
    }, [src]);

    useLayoutEffect(() => {
        if (!loaded || !editorRef.current) {
            return;
        }

        const { cropper } = editorRef.current;

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

    useLayoutEffect(() => {
        if (!loaded || !editorRef.current) {
            return;
        }

        const { cropper } = editorRef.current;

        /**
         * Rotate the image.
         */
        cropper.rotateTo(-rotate);
    }, [loaded, rotate]);

    useImperativeHandle(ref, () => editorRef.current!);

    return (
        <>
            <div className={cx("editor")}>
                <Cropper
                    ref={editorRef}
                    className={cx("cropper")}
                    src={src}
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
                    cropstart={onEditingStart}
                    cropend={onEditingEnd}
                    onLoad={() => {
                        setLoaded(true);
                    }}
                    onError={event => {
                        if (onError) {
                            onError(event);
                        }
                    }}
                />

                <div className={cx("zoom-controls")}>
                    <div className={cx("zoom")}>
                        <Button
                            className={cx("zoom-out-button")}
                            size="icon-sm"
                            variant="link"
                            disabled={zoom === 0}
                            onClick={() => {
                                setZoom(prevZoom => Math.max(prevZoom - 10, 0));
                            }}
                        >
                            <ZoomOut className={cx("zoom-icon")} />
                        </Button>
                        <div className={cx("zoom-slider")}>
                            <Slider
                                step={1}
                                min={0}
                                max={100}
                                value={[zoom]}
                                onValueChange={vals => {
                                    setZoom(vals[0] ?? 0);
                                }}
                                onFocus={onEditingStart}
                                onBlur={onEditingEnd}
                            />
                            <div className="text-xs text-muted-foreground mt-1 text-center">{zoomFormatter(zoom)}</div>
                        </div>
                        <Button
                            className={cx("zoom-in-button")}
                            size="icon-sm"
                            variant="link"
                            disabled={zoom === 100}
                            onClick={() => {
                                setZoom(prevZoom => Math.min(prevZoom + 10, 100));
                            }}
                        >
                            <ZoomIn className={cx("zoom-icon")} />
                        </Button>
                    </div>

                    <div className="flex gap-2">
                        {SHOW_PREVIEW && (
                            <Button
                                className={cx("preview-button")}
                                size="icon-sm"
                                variant="outline"
                                onClick={() => {
                                    const editor = editorRef.current?.cropper;

                                    if (!editor) {
                                        return;
                                    }

                                    const canvas = editor.getCroppedCanvas({
                                        width: 1024,
                                        height: 1024,
                                    });

                                    setPreviewOpen({
                                        open: true,
                                        src: canvas.toDataURL("image/jpeg", 0.9),
                                    });
                                }}
                            >
                                <Eye className={cx("preview-icon")} />
                            </Button>
                        )}

                        <Button
                            className={cx("reset-button")}
                            size="icon-sm"
                            variant="outline"
                            onClick={() => {
                                editorRef.current?.cropper.reset();

                                setZoom(0);
                                setRotate(0);
                            }}
                        >
                            <RotateCcw className={cx("reset-icon")} />
                        </Button>
                    </div>
                </div>

                <div className={cx("rotate-controls")}>
                    <div className={cx("rotate-slider")}>
                        <div className="flex h-full flex-col items-center justify-between pb-2 pt-2">
                            <span className="text-xs text-muted-foreground">180°</span>
                            <Slider
                                orientation="vertical"
                                step={1}
                                min={-180}
                                max={180}
                                value={[rotate]}
                                onValueChange={vals => {
                                    setRotate(vals[0] ?? 0);
                                }}
                                onFocus={onEditingStart}
                                onBlur={onEditingEnd}
                                className="h-[200px]"
                            />
                            <span className="text-xs text-muted-foreground">0°</span>
                        </div>
                        <div className="absolute left-0 top-1/2 -translate-x-full -translate-y-1/2 whitespace-nowrap text-xs text-muted-foreground">
                            {rotateFormatter(rotate)}
                        </div>
                    </div>
                </div>
            </div>

            <Dialog
                open={previewOpen.open}
                onOpenChange={open => {
                    if (!open) {
                        setPreviewOpen({
                            open: false,
                            src: undefined,
                        });
                    }
                }}
            >
                <DialogContent className="sm:max-w-2xl">
                    <DialogHeader>
                        <DialogTitle>Preview</DialogTitle>
                        <DialogDescription>Preview of cropped image</DialogDescription>
                    </DialogHeader>
                    {previewOpen.src && (
                        <div className="flex items-center justify-center">
                            <img
                                src={previewOpen.src}
                                alt="Preview"
                                className="max-w-full rounded-lg"
                            />
                        </div>
                    )}
                </DialogContent>
            </Dialog>
        </>
    );
}

interface Props {
    ref?: React.Ref<ReactCropperElement | HTMLImageElement>;
    src: string;
    onError?: React.ReactEventHandler<HTMLImageElement> | undefined;
    onEditingStart?: () => void;
    onEditingEnd?: () => void;
}

export const Editor = React.memo(View);
