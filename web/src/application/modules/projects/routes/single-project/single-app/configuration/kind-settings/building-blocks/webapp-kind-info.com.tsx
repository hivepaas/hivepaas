import { Info } from "lucide-react";

export function WebappKindInfo() {
    return (
        <div className="flex items-start gap-3 rounded-lg border border-border bg-muted/30 p-4 max-w-[800px]">
            <Info className="size-5 text-muted-foreground shrink-0 mt-0.5" />
            <div className="flex flex-col gap-1 text-sm">
                <span className="font-medium text-foreground">Standard Web Application</span>
                <p className="text-muted-foreground text-xs leading-relaxed">
                    Web applications serve HTTP traffic and microservices. Detailed networking rules, custom domains,
                    SSL certificates, and health checks are configured in{" "}
                    <span className="font-medium text-foreground">Routing Settings</span> and{" "}
                    <span className="font-medium text-foreground">Container Settings</span>.
                </p>
            </div>
        </div>
    );
}
