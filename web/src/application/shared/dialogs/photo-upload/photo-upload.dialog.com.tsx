import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogActionFooter,
    DialogDescription,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";

import { UploadPhotoForm } from "@application/shared/form";

interface PhotoUploadDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSubmit: (result: File | null) => void;
    initialImage?: string | null;
    title?: string;
    description?: string;
    filename?: string;
}

export function PhotoUploadDialog({
    open,
    onOpenChange,
    onSubmit,
    initialImage = null,
    title = "Profile Photo",
    description = "Adjust the crop area. Use tools to rotate/zoom.",
    filename,
}: PhotoUploadDialogProps) {
    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
            modal
        >
            <DialogFixedContent className="sm:max-w-2xl">
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                    <DialogDescription>{description}</DialogDescription>
                </DialogHeader>

                <UploadPhotoForm
                    photo={initialImage}
                    filename={filename}
                    onSubmit={result => {
                        onSubmit(result.photo);
                        onOpenChange(false);
                    }}
                    footer={
                        <DialogActionFooter>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => {
                                    onOpenChange(false);
                                }}
                            >
                                Cancel
                            </Button>
                            <Button
                                type="submit"
                                className="min-w-[100px]"
                            >
                                Save
                            </Button>
                        </DialogActionFooter>
                    }
                />
            </DialogFixedContent>
        </Dialog>
    );
}
