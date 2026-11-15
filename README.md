# Jaguar Desktop Pet — Windows CI & Store validation (no local admin required)

This source bundle is for **GitHub-hosted Windows builds**, with no need to execute unsigned code on a company-managed PC. It is based on the v0.1 Go Win32 transparent-jaguar prototype and the identity shown in Partner Center.

## What is tested automatically
- Run motion unit tests and `go vet` on Microsoft's Windows environment hosted via GitHub Actions `windows-2022`.
- Compile a Windows console test binary and execute `--ci-selftest` (all 8 distinct PNG frames, premultiplied pixel buffers and motion update).
- Compile the actual Windows GUI `.exe` from the same source.
- Build an **official MakeAppx unsigned MSIX**, unpack it, and check the app identity, publisher, architecture and executable presence.
- Upload **an unsigned Store package as a GitHub Actions artifact**, plus a result text file.

These steps have **not yet run on GitHub's Windows servers**. The workflow is prepared but must be uploaded to a repository and executed before these checks are considered passed.

## Browser-only setup: GitHub Actions
1. Visit https://github.com/new and create a **private** repository, e.g. `jaguar-desktop-pet` (private repository minutes are subject to your plan). Prefer not to store employer materials in it.
2. Upload the entire **contents of this folder**, including `.github/workflows/windows-msix.yml`, `app`, `packaging`, `scripts`, and `store-assets`. Do not upload only the outer ZIP as a single repo file. The workflow file must remain at `.github/workflows/windows-msix.yml` on the default branch `main`.
3. Open the repository **Actions** tab; authorize Actions for the repository if requested and permitted by your GitHub account. Select `Windows build and MSIX validation`, then **Run workflow**.
4. Inspect all workflow steps. If green, download the `jaguar-store-candidate-unsigned-*` artifact ZIP and extract the `.msix` inside. **Do not install it on your managed work PC**.
5. Before any Store submission, perform interactive UI testing on a separately authorized Cloud PC or Windows computer. Validate visual overlay, animation, tray controls, pause, and exit. Obtain real screenshots of that working app. The GitHub runner **is not an interactive display test**.
6. After a passing UI test and the required real Store screenshots and privacy URL, choose **Private audience** under Partner Center > Pricing and availability, add testers' **personal** Microsoft accounts, upload the final unsigned MSIX, and submit to Microsoft for certification. Do not publish as Public audience yet. Store signing happens **after certification**.
7. If employer policy explicitly allows installing this Store-distributed private app, test the certified version through Microsoft Store. Otherwise use a personally controlled or IT-authorized machine. A Store-signed package does not waive corporate controls.

## Alternative for real GUI test without a local PC
Use a Windows 365 Cloud PC, if you already have a legitimately provisioned license and permission, via https://windows.cloud.microsoft/ in your browser. This is a real interactive Windows session, unlike GitHub Actions. It requires a separate Cloud PC account/license. You may also use an approved Windows VM or an authorized tester's PC. No changes to your work PC's antivirus are needed.

## Limitations and release gates
- **No claim of Windows GUI success yet.** Previous local attempts did not produce a working desktop jaguar.
- CI tests validate code and packaging, but cannot prove that the Win32 overlay is visible or responsive on an interactive desktop.
- GitHub Actions package is unsigned and is **only an upload candidate**, not a verified secure installer, not an approved final release, not a 3D rigged jaguar.
- WACK has been deprecated per updated Microsoft documentation; optional local tests may still be useful, but official Store certification is the authoritative external check.
- Never change device policy, disable Defender, bypass SmartScreen, or install unsigned software on a work computer without authorization.

## Store identity
Package Name: `SamerTallauze.JaguarDesktopPet`  
Publisher: `CN=C83AFC34-5CE3-4E0A-8DCD-9733097BE6AC`  
Publisher Display Name: `Samer Tallauze`  
Version: `0.1.0.0` (development)  
Store ID: `9NNHDHM3WGQ9`

Sources:
- GitHub hosted Windows runners: https://docs.github.com/en/actions/how-tos/write-workflows/choose-where-workflows-run/choose-the-runner-for-a-job
- MakeAppx SDK: https://learn.microsoft.com/en-us/windows/msix/package/create-app-package-with-makeappx-tool
- Private audience: https://learn.microsoft.com/en-us/windows/apps/publish/publish-your-app/msix/visibility-options
- Store certification: https://learn.microsoft.com/en-us/windows/apps/publish/publish-your-app/msix/app-certification-process
- Windows 365 browser access: https://learn.microsoft.com/en-us/windows-365/end-user-access-cloud-pc
