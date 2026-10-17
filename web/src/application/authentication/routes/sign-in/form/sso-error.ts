/**
 * What the sign-in page says when a sign-in with a provider failed: the server
 * sends the browser back with the error's code (`?ssoError=`), never its text.
 */
const SSO_ERROR_MESSAGES: Record<string, string> = {
    ERR_OAUTH_EMAIL_NOT_VERIFIED:
        "The provider does not confirm that your email address is verified. Verify it with the provider, then sign in again.",
    ERR_OAUTH_USER_NOT_FOUND:
        "No user has the email address you signed in with. Ask an administrator to add you, or sign in another way.",
    ERR_OAUTH_USER_EMAIL_NOT_RETURNED:
        "The provider did not share your email address. Ask an administrator to check the sign-in's scopes and permissions.",
    ERR_USER_UNAVAILABLE: "Your account cannot sign in. Contact an administrator.",
    ERR_UNAUTHORIZED: "The sign-in with the provider was not completed. Try again.",
    ERR_NOT_FOUND: "This way of signing in is not available any more. Choose another.",
};

const SSO_ERROR_FALLBACK = "The sign-in with the provider failed. Try again, or sign in another way.";

export function ssoErrorMessage(code: string): string {
    return SSO_ERROR_MESSAGES[code] ?? SSO_ERROR_FALLBACK;
}
