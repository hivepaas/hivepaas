import { cn } from "@lib/utils";
import { Link } from "react-router";

interface Tile {
    label: string;
    value: string;
    note?: string;
    /** Where the number comes from; a tile without one only informs. */
    to?: string;
    /** Something is wrong: the note is said louder. */
    alert?: boolean;
}

/** The few numbers worth seeing first, each leading to where they come from. */
export function SummaryTiles({ tiles }: { tiles: Tile[] }) {
    return (
        <section
            aria-label="Summary"
            className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4"
        >
            {tiles.map(tile => {
                const body = (
                    <>
                        <span className="text-[13px] font-medium text-muted-foreground">{tile.label}</span>
                        <span className="font-mono text-3xl font-bold tracking-tight">{tile.value}</span>
                        {tile.note && (
                            <span
                                className={cn(
                                    "text-xs text-muted-foreground",
                                    tile.alert && "font-semibold text-destructive",
                                )}
                            >
                                {tile.note}
                            </span>
                        )}
                    </>
                );
                const className = "flex flex-col gap-2 rounded-xl border bg-card px-5 py-4";
                return tile.to ? (
                    <Link
                        key={tile.label}
                        to={tile.to}
                        className={cn(className, "hover:bg-muted/30")}
                    >
                        {body}
                    </Link>
                ) : (
                    <div
                        key={tile.label}
                        className={className}
                    >
                        {body}
                    </div>
                );
            })}
        </section>
    );
}
