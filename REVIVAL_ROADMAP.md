# Terraformer Revival — Roadmap (DRAFT, uncommitted)

Fork: https://github.com/Sun-Wukhan/terraformer
Upstream: GoogleCloudPlatform/terraformer (archived 2026-03-16, 14.5k stars, Apache-2.0)

## Where we are
- Fork created 2026-10-14, default branch `master`, not archived.
- Local working clone: `~/workspace/oss-contributions/terraformer`
- Repo settings: **Issues and Discussions are disabled by default on the fork**
  (GitHub API shows `has_issues: false`, `has_discussions: false`).
  Enable both in repo Settings if this revival wants community input.

## Tier 1 — make it work again (verified, scoped)
Goal: run current provider binaries (protocol 5 & 6) instead of the frozen
TF 0.12 gRPC client.

Verified today (2026-10-14):
- Only **13 files** import `github.com/hashicorp/terraform` (the v0.12.31 tree).
- The core is `terraformutils/providerwrapper/provider.go` (357 lines):
  `tfplugin.GRPCProvider` + `providers.GetSchemaResponse` + `configs/configschema`
  shelling out to user-installed provider binaries in `.terraform`.
- Replacement design: shell out to a modern `terraform`/`tofu` CLI —
  temp config + `terraform init` + `providers schema -json`, parse JSON.
  This deletes the whole 0.12 dependency tree and works with any
  registry-installed provider, protocol 5 or 6.
- Add OpenTofu support via the same path (tofu binary + OpenTofu registry).
- Auto-install provider binaries from the registry instead of requiring
  hand-placed binaries (pairs naturally with the CLI approach).

## Tier 2 — modernize the ecosystem
- Migrate Azure from dead track-1 SDK (`v63.4.0+incompatible`) to track-2
  (`sdk/resourcemanager/...`).
- Refresh AWS/GCP/IBM SDKs, bump Go toolchain (go.mod says 1.24.0), revive CI.
- Emit modern HCL: drop `${}` legacy interpolation, `required_providers`
  blocks, `terraform fmt`-clean output natively.

## Tier 3 — coverage and quality
- New resource coverage for services launched since ~2024 (big three clouds).
- Fix silent error-swallowing in resource enumeration
  (`log.Println(err)` + continue → surfaced errors).
- README banner announcing the community continuation + contributing notes.

## Suggested first commits (after tier pick)
1. README banner: "community-maintained continuation of the archived
   GoogleCloudPlatform/terraformer".
2. Enable Issues + Discussions in repo settings.
3. Scaffold GitHub Actions CI (go build/vet on push) — currently none runs.
