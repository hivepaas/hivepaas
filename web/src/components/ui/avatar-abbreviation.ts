/** The letters an avatar without a picture shows: the first of the name's first two words. */
export function avatarAbbreviation(name: string): string {
    return name
        .trim()
        .split(" ")
        .map(word => word[0]?.toUpperCase())
        .filter(Boolean)
        .slice(0, 2)
        .join("");
}
