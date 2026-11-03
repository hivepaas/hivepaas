import { memo, useState } from "react";

import { Button } from "@components/ui/button";
import { Copy } from "lucide-react";
import { toast } from "sonner";

type Props = {
    qrCode: string;
    secretKey: string;
};

function View({ qrCode, secretKey }: Props) {
    const [showSecret, setShowSecret] = useState(false);

    return (
        <>
            <img
                src={`data:image/png;base64,${qrCode}`}
                alt="QR Code"
                className="w-[350px] h-[350px] max-w-full aspect-square object-contain mx-auto"
            />
            <div className="flex items-center justify-center min-h-[36px]">
                {!showSecret ? (
                    <Button
                        variant="link"
                        size="sm"
                        type="button"
                        onClick={() => {
                            setShowSecret(true);
                        }}
                        className="text-sm text-muted-foreground hover:text-foreground"
                    >
                        Show code
                    </Button>
                ) : (
                    <div className="flex items-center justify-center gap-1">
                        <p className="text-sm font-mono text-muted-foreground text-center select-all">{secretKey}</p>
                        <Button
                            variant="link"
                            size="icon"
                            type="button"
                            onClick={e => {
                                e.preventDefault();
                                void navigator.clipboard.writeText(secretKey);
                                toast.success("Secret copied to clipboard");
                            }}
                        >
                            <Copy className="size-4" />
                        </Button>
                    </div>
                )}
            </div>
        </>
    );
}

export const MfaQrCode = memo(View);
