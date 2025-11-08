import { useMemo, useRef, useState } from "react";
import Cropper, { type ReactCropperElement } from "react-cropper";
import "cropperjs/dist/cropper.css";

import { Button } from "@components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@components/ui/dialog";

interface CroppedImageResult {
	fileName: string;
	dataBase64: string;
}

interface PhotoUploadDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onConfirm: (result: CroppedImageResult) => void;
	initialImage?: string | null;
	aspectRatio?: number;
}

export function PhotoUploadDialog({
	open,
	onOpenChange,
	onConfirm,
	initialImage = null,
	aspectRatio = 1,
}: PhotoUploadDialogProps) {
	const cropperRef = useRef<ReactCropperElement | null>(null);
	const [selectedFile, setSelectedFile] = useState<File | null>(null);
	const [imageSrc, setImageSrc] = useState<string | null>(initialImage);
	const [isCropping, setIsCropping] = useState(false);

	const fileName = useMemo(() => {
		return selectedFile?.name ?? "image.png";
	}, [selectedFile]);

	function onFileChange(e: React.ChangeEvent<HTMLInputElement>) {
		const file = e.target.files?.[0];

		if (file === undefined) {
			return;
		}

		const reader = new FileReader();
		reader.onload = () => {
			setSelectedFile(file);
			setImageSrc(typeof reader.result === "string" ? reader.result : null);
		};
		reader.readAsDataURL(file);
	}

	async function handleConfirm() {
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
		>
			<DialogContent className="sm:max-w-2xl">
				<DialogHeader>
					<DialogTitle>Upload and crop photo</DialogTitle>
					<DialogDescription>Select an image and adjust the crop area.</DialogDescription>
				</DialogHeader>

				<div className="flex flex-col gap-4">
					<input
						type="file"
						accept="image/*"
						onChange={onFileChange}
					/>

					{imageSrc && (
						<div className="w-full">
							<Cropper
								ref={cropperRef}
								style={{ height: 360, width: "100%" }}
								aspectRatio={aspectRatio}
								src={imageSrc}
								viewMode={1}
								guides
								background={false}
								autoCropArea={1}
								checkOrientation={false}
								responsive
							/>
						</div>
					)}
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
							void handleConfirm();
						}}
					>
						Save
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}


