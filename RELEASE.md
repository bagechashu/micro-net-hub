# Release Notes

Docker image: `xianleeti/micro-net-hub:<version>` (linux/amd64).

## v0.6.9 (2026-11-22)

Fix for v0.6.8 on MySQL: the `users.password` column was `varchar(255)`, but a password encrypted with a 2048-bit key is 344 characters long.

- `users.password` is now `varchar(1024)`. The normal startup migration widens existing columns automatically; existing data is kept.
- `rotate-rsa-key` checks the column first and widens it when it is too short (`--dry-run` only reports it), so rotation also works when run directly after upgrading.
- Fixes: on v0.6.8, `rotate-rsa-key` failed with `Data too long for column 'password'` (the transaction was rolled back, nothing was changed), and a fresh MySQL deployment could not create the initial admin or new users.

**Upgrade notes**

- If `rotate-rsa-key` failed on v0.6.8: nothing was changed, keep running as before. Upgrade to v0.6.9 and run the rotation again. The `rsa-backup-<time>/` directory left by the failed run only contains unused keys and can be deleted.

## v0.6.8 (2026-11-22)

RSA key handling: each deployment now uses its own key, and existing deployments can rotate away from the key that used to ship with this repository.

- The frontend fetches the RSA public key at runtime from the new public endpoint `GET /api/base/pubkey`. It is no longer baked into the build (`VITE_PUBLIC_KEY` is no longer used), so the same image works with any key.
- On startup the backend checks the RSA key files:
  - new deployment (no key files and no users in the database): a 2048-bit key pair is generated automatically;
  - users exist but the key files are missing (e.g. a volume was not mounted): the service refuses to start instead of generating a new key that would lock everyone out;
  - only one key file, or a public/private key mismatch: the service refuses to start;
  - still using the public key from this repository: a warning is logged.
- New command `micro-net-hub rotate-rsa-key [--dry-run]`: generates a new key pair and re-encrypts all stored user passwords and the encrypted passwords in `config.yml`. Users keep their passwords. Old keys, old config and the new keys are backed up to `rsa-backup-<time>/` before anything is changed; the database is updated in one transaction; `--dry-run` changes nothing.
- `ldap.admin-pass` and `ldap.user-init-password` in `config.yml` accept plain text with the `plain:` prefix.
- docker-compose examples no longer ship a private key: they mount `config/app/keys/` and the key pair is generated there on first start.

**Upgrade notes**

- Deployments that still use the repository key (the startup warning tells you): upgrade to this image, stop the service, run `rotate-rsa-key --dry-run`, then `rotate-rsa-key`, then start. See [README → RSA key and key rotation](README.md#rsa-key-and-key-rotation).
- New docker-compose deployments: `sudo chown 1000:1000 config/app/keys` before the first start, and back up that directory.
- If you build the frontend yourself, you can delete `VITE_PUBLIC_KEY` from `frontend/.env.*`.

## v0.6.7 (2026-11-22)

- Removed the unmaintained Vue 2 frontend (`vue2/`). The Vue 3 UI in `frontend/` is the only frontend. No change to the application itself.

## v0.6.6 (2026-11-22)

- Migrated to casbin v3 (`github.com/casbin/casbin/v3` v3.11.0) and `casbin/gorm-adapter` v3.41.0. Behaviour was verified to be identical to v2 for all policy operations the project uses, and against the existing policies in a real database.
- Fixed API seed data: wrong remarks and categories for the DNS APIs, `GET /dns/getall` (was recorded as POST), `/sitenav/getnav` (was recorded as a non-existent `/sitenav/list`). The casbin seed policies were updated to match.
- The sample nav site uses the `si:jenkins` brand icon instead of a removed image.
- A fresh install no longer logs an error when no DingTalk/WeCom/Feishu groups are configured.
- `docs/sql/micro_net_hub.sql` regenerated from a fresh initialisation with the current code (`utf8mb4_general_ci`, imports on MySQL 5.7).

**Upgrade notes**

- Seed data is only written on a fresh install. Existing databases keep the old API records; fix them in "API management" if needed.

## v0.6.5 (2026-11-22)

- New "Session expired" dialog. "Sign in again" opens the login dialog and returns to the previous page after login.
- Fixed unstyled confirm dialogs (e.g. batch delete): the message and message-box styles are now loaded globally.
- `system.site-logo` only replaces the logo in the page header; the browser tab icon stays the default favicon.

## v0.6.4 (2026-11-21)

- Site navigation bookmarks: bookmark sites from a card's menu; bookmarked sites are shown in a "Bookmarks" group at the top. Stored in the browser (localStorage).
- Configurable site name and logo: `system.site-name` and `system.site-logo` in `config.yml` (served by the public endpoints `GET /api/base/siteinfo` and `GET /api/base/sitelogo`). Emails fall back to the site name when `notice.project-name` is not set.

**Upgrade notes**

- Optional new config keys `system.site-name` and `system.site-logo`. A custom logo file has to be mounted into the container and readable by uid 1000.

## v0.6.3 (2026-11-21)

- Fixed a crash when saving a nav site without a document link.
- Redesigned 404 / 401 pages inside the main layout.
- Notification emails follow the system light / dark theme.

## v0.6.2 (2026-11-20)

- Built with Go 1.27; all Go dependencies upgraded. This fixes the critical and high CVEs Docker Scout reported in the Go standard library, `x/crypto`, `x/net`, `x/text`, `pgx` and `quic-go`.
- Docker image based on `alpine:3.24`, runs as the non-root user `app` (uid/gid 1000), exec-form `CMD`, `EXPOSE 9000`.

**Upgrade notes**

- The container no longer runs as root: mounted `config.yml` and RSA key files must be readable by uid 1000, and a mounted log directory must be writable by uid 1000.

## v0.6.1 (2026-11-20)

- Redesigned notification emails (account created, password reset, verification code, account status). Rendered with `html/template`, so all dynamic values are escaped.
- The TOTP QR code is sent as an inline image attachment, so Gmail and Outlook show it.
- Account notice files saved to disk use a sanitised file name.

## v0.6.0 (2026-11-20)

- New admin UI: the frontend was migrated from Vue 2 to Vue 3 (Vite, TypeScript, Element Plus, Pinia) with a Tabler-style redesign:
  - top navigation, light / dark / system theme, Tabler icons and a new logo;
  - restyled tables, dialogs and forms; profile page and password-reset page redesigned;
  - nav site icons: brand logos, Tabler icons, site favicon or image URL (all SVG).
- Removed the built-in LDAP service. Use an external OpenLDAP (see `docs/docker-compose/openldap`).
- Also includes earlier work: the `ocserv-users-web` tool (login history page, nft rules), and a `radius-client` CLI for testing RADIUS authentication.

**Upgrade notes**

- The old built-in nav icons under `/ui/assets/logo/` were removed. Sites still pointing to them show the first letter of the site name; pick a new icon in the nav manager.
