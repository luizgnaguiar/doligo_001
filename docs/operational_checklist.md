# Operational Checklist (Pre-GoLive)

This document outlines the minimum verifiable operational checklist for deploying the Doligo ERP/CRM system in non-massive production environments. This checklist is focused on operational aspects, infrastructure, and ensuring the system's basic functionality outside of code changes.

**Note:** This phase explicitly prohibits any code modifications. Identified technical needs should be registered as technical debt.

## 1. Environment Variables Documentation

*   **Status:** VERIFIED
*   **Verification:** `docs/env.md` has been created, listing all environment variables with clear descriptions, indicating their mandatory/optional status, and specifying default values.
*   **Evidence:** `docs/env.md`

## 2. Real Linux Binary Execution

*   **Status:** ANNOTATED (Placeholder for actual verification)
*   **Verification:** The application binary **MUST** be executed at least once in a real Linux environment (e.g., VM, Docker container, CI/CD pipeline agent running Linux). This verification includes:
    *   Successful process initialization.
    *   Successful connection to the configured database.
    *   Successful initialization of the HTTP server, making endpoints accessible.
*   **Evidence (Placeholder):** To be provided by an operational log, CI/CD pipeline output, or a manual annotation confirming the execution and its successful outcome in a real Linux environment.
    *   _Example Annotation:_ "On `[DATE]`, the `doligo` binary (version `[VERSION/COMMIT_HASH]`) was successfully executed on a `[Linux Distribution/Container Type]` environment. Process initialized, connected to PostgreSQL, and HTTP server started on port 8080."

## 3. Reverse Proxy / Gateway Configuration

*   **Status:** PENDING (Requires infrastructure team action)
*   **Verification:** A reverse proxy (e.g., Nginx, Traefik, Caddy) **MUST** be configured in the production environment.
    *   Request timeout for routes generating PDF documents (`/api/v1/invoices/{id}/pdf`) **MUST** be explicitly set to at least `30 seconds`.
*   **Evidence:** Recommendations documented in `docs/production_notes.md`. Configuration files of the reverse proxy (e.g., `nginx.conf`, Traefik dynamic configuration) are infrastructure-level and reside outside the application's repository.

## 4. Automatic Database Backup

*   **Status:** PENDING (Requires infrastructure team action)
*   **Verification:** An automated routine for database backups **MUST** be in place.
    *   A minimum retention policy (e.g., last 7 days, last 30 backups) **MUST** be defined and enforced.
*   **Evidence:** Documentation or configuration of the backup solution (e.g., cron jobs, cloud provider backup services). This is an operational responsibility.

## 5. Log Rotation on Host

*   **Status:** PENDING (Requires infrastructure team action)
*   **Verification:** The host executing the application **MUST** have log rotation configured (e.g., `logrotate` for Linux systems).
*   **Evidence:** Host-level log rotation configuration files (e.g., `/etc/logrotate.d/doligo`). This is an infrastructure-level configuration.

---
**Technical Debt Identified (Not Implemented in this Phase):**
*   _None in this specific phase, as per instructions. Any future needs will be documented here._
