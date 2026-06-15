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
            <DialogFixedContent className="w-[640px] max-w-[calc(100vw-2rem)]">
                <DialogHeader>
                    <DialogTitle className="text-2xl">{title}</DialogTitle>
                </DialogHeader>

                <DialogBody className="py-8">
                    <div className="grid grid-cols-[minmax(0,1fr)_auto] gap-x-8 gap-y-7 text-base sm:text-lg">
                        {rows.map(row => (
                            <div
                                key={row.label}
                                className="contents"
                            >
                                <span className="font-semibold">{row.label}</span>
                                <span>{row.value}</span>
                            </div>
                        ))}
                    </div>
                </DialogBody>

                <DialogActionFooter className="flex justify-end">
                    <Button
                        type="button"
                        className="min-w-[120px]"
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
