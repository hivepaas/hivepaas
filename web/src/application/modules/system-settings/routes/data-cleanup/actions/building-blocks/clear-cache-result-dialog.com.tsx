import { Button } from "@components/ui";
import {
    Dialog,
    DialogActionFooter,
    DialogBody,
    DialogFixedContent,
    DialogHeader,
    DialogTitle,
} from "@components/ui/dialog";

interface ClearCacheResultDialogProps {
    open: boolean;
    title: string;
    rows: ClearCacheResultRow[];
    onOpenChange: (open: boolean) => void;
}

export interface ClearCacheResultRow {
    label: string;
    value: string | number;
}

export function ClearCacheResultDialog({ open, title, rows, onOpenChange }: ClearCacheResultDialogProps) {
    return (
        <Dialog
            open={open}
            onOpenChange={onOpenChange}
        >
            <DialogFixedContent className="w-[400px] max-w-[calc(100vw-2rem)]">
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                </DialogHeader>

                <DialogBody className="">
                    <div className="flex flex-col gap-2">
                        {rows.map(row => (
                            <div
                                key={row.label}
                                className="flex items-center gap-3"
                            >
                                <div className="text-sm font-medium w-[145px]">{row.label}</div>
                                <div className="flex-1">{row.value}</div>
                            </div>
                        ))}
                    </div>
                </DialogBody>

                <DialogActionFooter className="flex justify-end">
                    <Button
                        type="button"
                        className="min-w-[100px]"
                        onClick={() => {
                            onOpenChange(false);
                        }}
                    >
                        Close
                    </Button>
                </DialogActionFooter>
            </DialogFixedContent>
        </Dialog>
    );
}
