import { expect, test } from "../../support/fixtures";

// Every page the sidebar reaches, with the heading it shows once loaded. A
// page added to the dashboard belongs here.
const PAGES: [path: string, heading: string][] = [
    ["/cluster/networks/", "Networks"],
    ["/cluster/networks/create/", "Create a network"],
    ["/cluster/nodes/", "Nodes"],
    ["/cluster/volumes/", "Volumes"],
    ["/cluster/volumes/create/", "Create a volume"],
    ["/current-user/api-keys/", "Your account"],
    ["/current-user/api-keys/create/", "Create a new API key"],
    ["/current-user/profile/", "Your account"],
    ["/home/", "Welcome back"],
    ["/integrations/access-tokens/", "Access Tokens"],
    ["/integrations/access-tokens/create/", "Create Access Token"],
    ["/integrations/acme-dns-providers/", "ACME DNS Providers"],
    ["/integrations/acme-dns-providers/create/", "Create ACME DNS Provider"],
    ["/integrations/backup-repos/", "Backup Repos"],
    ["/integrations/backup-repos/create/", "Create Backup Repo"],
    ["/integrations/backup-snapshots/", "Backup Snapshots"],
    ["/integrations/basic-auth/", "Basic Auth"],
    ["/integrations/basic-auth/create/", "Create Basic Auth"],
    ["/integrations/cloud-storages/", "Cloud Storages"],
    ["/integrations/cloud-storages/create/", "Create Cloud Storage"],
    ["/integrations/email-accounts/", "Email Accounts"],
    ["/integrations/email-accounts/create/", "Create Email Account"],
    ["/integrations/github-apps/", "Github Apps"],
    ["/integrations/github-apps/create/", "Create Github App"],
    ["/integrations/im-platforms/", "IM Platforms"],
    ["/integrations/im-platforms/create/", "Create IM Platform"],
    ["/integrations/key-auth/", "Key Auth"],
    ["/integrations/key-auth/create/", "Create Key Auth"],
    ["/integrations/notification-targets/", "Notification Targets"],
    ["/integrations/notification-targets/create/", "Create Notification Target"],
    ["/integrations/oauth/", "OAuth"],
    ["/integrations/oauth/create/", "Create OAuth"],
    ["/integrations/registry-auth/", "Registry Auth"],
    ["/integrations/registry-auth/create/", "Create Registry Auth"],
    ["/integrations/ssh-keys/", "SSH Keys"],
    ["/integrations/ssh-keys/create/", "Create SSH Key"],
    ["/integrations/ssl-certificates/", "SSL Certificates"],
    ["/integrations/ssl-certificates/create/", "Create SSL Certificate"],
    ["/integrations/ssl-providers/", "SSL Providers"],
    ["/integrations/ssl-providers/create/", "Create SSL Provider"],
    ["/integrations/webhooks/", "Webhooks"],
    ["/integrations/webhooks/create/", "Create Webhook"],
    ["/operations/audit-logs/", "Audit Logs"],
    ["/operations/export/", "Export"],
    ["/operations/tasks/", "Tasks"],
    ["/projects/", "Projects"],
    ["/projects/new/compose/", "Projects"],
    ["/settings/app-placement/", "App Placement"],
    ["/settings/backup-repo-cleanup/actions/", "Backup Repo Cleanup"],
    ["/settings/backup-repo-cleanup/configuration/", "Backup Repo Cleanup"],
    ["/settings/data-backup/actions/", "Data Backup"],
    ["/settings/data-backup/configuration/", "Data Backup"],
    ["/settings/data-backup/snapshots/", "Data Backup"],
    ["/settings/data-cleanup/actions/", "Data Cleanup"],
    ["/settings/data-cleanup/configuration/", "Data Cleanup"],
    ["/settings/image-build/", "Image Build"],
    ["/settings/registry-auth-renewal/actions/", "Registry Auth Renewal"],
    ["/settings/registry-auth-renewal/configuration/", "Registry Auth Renewal"],
    ["/settings/ssl-renewal/actions/", "SSL Renewal"],
    ["/settings/ssl-renewal/configuration/", "SSL Renewal"],
    ["/system/ai/mcp/", "AI"],
    ["/system/hivepaas/actions/", "HivePaaS"],
    ["/system/hivepaas/general/", "HivePaaS"],
    ["/system/hivepaas/routing-settings/", "HivePaaS"],
    ["/system/hivepaas/security/", "HivePaaS"],
    ["/system/hivepaas/updates/", "HivePaaS"],
    ["/system/logging/configuration/", "Logging"],
    ["/system/logging/routes-and-calls/", "Logging"],
    ["/system/registry/configuration/", "Registry"],
    ["/system/traefik/actions/", "Traefik"],
    ["/system/traefik/config-options/", "Traefik"],
    ["/system/traefik/general/", "Traefik"],
    ["/user-management/users/", "Users"],
];

// Each page opens as an admin sees it: its heading shows, it stays at its
// address, and nothing went wrong on the way - no error toast, no error in the
// console, no API call refused or failed.
for (const [path, heading] of PAGES) {
    test(`${path} opens cleanly`, async ({ page }) => {
        const problems: string[] = [];
        page.on("console", msg => {
            if (msg.type() === "error") problems.push(`console: ${msg.text()}`);
        });
        page.on("response", res => {
            if (res.url().includes("/api/") && res.status() >= 400) {
                problems.push(`${res.status()} ${res.request().method()} ${res.url()}`);
            }
        });

        await page.goto(path);

        await expect(page.getByRole("heading", { name: heading }).first()).toBeVisible();
        await page.waitForLoadState("networkidle");
        expect(new URL(page.url()).pathname).toBe(path);
        await expect(page.locator('[data-sonner-toast][data-type="error"]')).toHaveCount(0);
        expect(problems).toEqual([]);
    });
}
