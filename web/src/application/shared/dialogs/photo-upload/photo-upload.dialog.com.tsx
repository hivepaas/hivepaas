import { Button } from "@components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";

import { UploadPhotoForm } from "@application/shared/form";

interface PhotoUploadDialogProps {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSubmit: (result: File | null) => void;
    initialImage?: string | null;
}

export function PhotoUploadDialog({ open, onOpenChange, onSubmit, initialImage = null }: PhotoUploadDialogProps) {
    console.log(open);
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

                <UploadPhotoForm
                    photo={initialImage}
                    onSubmit={result => {
                        onSubmit(result.photo);
                        onOpenChange(false);
                    }}
                >
                    <DialogFooter className="mt-5">
                        <Button
                            variant="outline"
                            onClick={() => {
                                onOpenChange(false);
                            }}
                        >
                            Cancel
                        </Button>
                        <Button type="submit">Save</Button>
                    </DialogFooter>
                </UploadPhotoForm>
            </DialogContent>
        </Dialog>
    );
}
