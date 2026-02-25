# Testing Deployment — K8s QA Environment

**Date:** 2026-02-25
**Branch:** deploy/testing-qa

## Summary

Deployed YaaiCMS to the K3S testing cluster for human QA.

## Changes

- **Ingress:** Changed base `ingressClassName` from `nginx` to `traefik` (matches K3S cluster). Removed nginx-specific annotation.
- **Container Image:** Updated deployment to pull from in-cluster `private-registry.default.svc.cluster.local:5000/yaaicms:latest`.
- **Valkey Port:** Fixed port from 6379 to 6380 (correct port on the testing cluster).
- **Seeding:** Extended database seeding to also run in `testing` environment (was dev-only).

## Infrastructure

- **Namespace:** `yaaicms`
- **Registry:** Used in-cluster `private-registry` (port-forward from dev machine to push images)
- **TLS:** cert-manager with `letsencrypt-prod` ClusterIssuer
- **Database:** PostgreSQL at 10.0.0.6:5432 — created `testing_smartpress` user and database
- **Cache:** Valkey at 10.0.0.6:6380
- **URL:** https://yaaicms.test.vlah.sh

## Verification

- Pod running (1/1 Ready)
- TLS certificate provisioned and Ready
- Homepage, admin login, blog post, health endpoint all returning HTTP 200
- Database seeded: admin user, 4 templates, 2 content items
- AI providers initialized (gemini active, all 4 available)
- S3 storage connected

## QA Login

- **URL:** https://yaaicms.test.vlah.sh/admin/login
- **Email:** admin@yaaicms.local
- **Password:** admin
- 2FA setup required on first login
