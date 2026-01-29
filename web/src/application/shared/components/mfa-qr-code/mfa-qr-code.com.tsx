import { memo } from "react";

import { Button } from "@components/ui/button";
import { Copy } from "lucide-react";
import { toast } from "sonner";

type Props = {
    qrCode: string;
    secretKey: string;
};

function View({ qrCode, secretKey }: Props) {
    return (
        <>
            <img
                src={`data:image/png;base64,${qrCode}`}
                alt="QR Code"
                className="w-[300px] h-[300px] object-contain"
            />
            <div className="flex items-center justify-center">
                <p className="text-sm text-gray-500 text-center">{secretKey}</p>
                <Button
                    variant="link"
                    size="icon"
                    onClick={e => {
                        e.preventDefault();
                        void navigator.clipboard.writeText(secretKey);
                        toast.success("Secret copied to clipboard");
                    }}
                >
                    <Copy className="size-4" />
                </Button>
            </div>
        </>
    );
}

export const MfaQrCode = memo(View);
