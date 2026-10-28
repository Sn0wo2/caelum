const { appendFileSync, closeSync, cpSync, createReadStream, mkdirSync, openSync, statSync, writeFileSync } = require("node:fs")
const { createHash } = require("node:crypto")
const { join } = require("node:path")
const { spawnSync } = require("node:child_process")

function requiredEnv(name) {
  const value = process.env[name]
  if (!value) throw new Error(`${name} is not set`)
  return value
}

function raw(value) {
  if (value === null || value === undefined) return "null"
  if (typeof value === "object") return JSON.stringify(value)
  return String(value)
}

function ghApi(endpoint) {
  const result = spawnSync("gh", ["api", endpoint], {
    encoding: "utf8",
    maxBuffer: 64 * 1024 * 1024,
    stdio: ["ignore", "pipe", "inherit"],
  })
  if (result.error) throw result.error
  if (result.status !== 0) throw new Error(`gh api exited with status ${result.status}`)
  return result.stdout
}

function commandToFile(command, args, file) {
  const fd = openSync(file, "w")
  let result
  try {
    result = spawnSync(command, args, { stdio: ["ignore", fd, "inherit"] })
  } finally {
    closeSync(fd)
  }
  if (result.error) throw result.error
  if (result.status !== 0) throw new Error(`${command} exited with status ${result.status}`)
}

function annotatedError(message) {
  throw new Error(`::error::${message}`)
}

async function sha256(file) {
  const hash = createHash("sha256")
  for await (const chunk of createReadStream(file)) hash.update(chunk)
  return hash.digest("hex")
}

async function main() {
  requiredEnv("GH_TOKEN")
  const actaTag = requiredEnv("ACTA_TAG")
  const runnerTemp = requiredEnv("RUNNER_TEMP")
  const workspace = requiredEnv("GITHUB_WORKSPACE")
  const githubOutput = requiredEnv("GITHUB_OUTPUT")

  const releaseEndpoint = actaTag === "latest"
    ? "repos/Sn0wo2/acta/releases/latest"
    : `repos/Sn0wo2/acta/releases/tags/${actaTag}`
  const release = JSON.parse(ghApi(releaseEndpoint))
  const tag = raw(release.tag_name)
  const url = raw(release.html_url)
  if (actaTag === "latest" && (release.draft === true || release.prerelease === true)) {
    annotatedError(`Acta release ${tag} is not an eligible stable release`)
  }

  const version = tag.startsWith("v") ? tag.slice(1) : tag
  const archives = [
    `acta_${version}_ffi_linux-windows.tar.gz`,
    `acta_${version}_ffi_macos.tar.gz`,
  ]
  const assets = Array.isArray(release.assets) ? release.assets : []
  const available = assets.map((asset) => raw(asset.name)).join(", ")
  const archiveAssets = new Map()
  for (const archive of archives) {
    const asset = assets.find((candidate) => candidate.name === archive)
    if (!asset) {
      annotatedError(`Acta release ${tag} is missing required FFI archive ${archive}. Available assets: ${available || "none"}. No native files were changed.`)
    }
    archiveAssets.set(archive, asset)
  }

  const digests = new Map()
  const withoutDigest = []
  for (const archive of archives) {
    const digest = raw(archiveAssets.get(archive).digest)
    const match = /^sha256:([a-f0-9]{64})$/.exec(digest === "null" ? "" : digest)
    if (match) digests.set(archive, match[1])
    else withoutDigest.push(archive)
  }

  if (withoutDigest.length > 0) {
    const checksumAsset = assets.find((asset) => asset.name === "checksums.txt")
    const checksumsUrl = checksumAsset ? raw(checksumAsset.browser_download_url) : ""
    if (!checksumsUrl) {
      annotatedError(`Acta release ${tag} has no SHA256 digest for every FFI archive and no checksums.txt. No native files were changed.`)
    }
    const checksums = ghApi(checksumsUrl)
    for (const archive of withoutDigest) {
      let digest = ""
      for (const line of checksums.split(/\r?\n/)) {
        const fields = line.trim().split(/\s+/)
        if (fields.length >= 2 && (fields[1] === archive || fields[1] === `*${archive}`)) {
          digest = fields[0]
          break
        }
      }
      if (!/^[a-f0-9]{64}$/.test(digest)) {
        annotatedError(`Acta checksums.txt has no SHA256 digest for ${archive}. No native files were changed.`)
      }
      digests.set(archive, digest)
    }
  }

  const archivePaths = new Map()
  for (const archive of archives) {
    const assetUrl = raw(archiveAssets.get(archive).browser_download_url)
    const archivePath = join(runnerTemp, archive)
    commandToFile("gh", ["api", assetUrl], archivePath)
    if (await sha256(archivePath) !== digests.get(archive)) {
      annotatedError(`SHA256 mismatch for Acta archive ${archive}`)
    }
    archivePaths.set(archive, archivePath)
  }

  const staged = join(runnerTemp, "staged")
  const [linuxArchive, macosArchive] = archives
  const mappings = [
    [linuxArchive, "x86_64-unknown-linux-gnu/libacta_ffi.so", "linux_amd64/libacta_ffi.so"],
    [linuxArchive, "aarch64-unknown-linux-gnu/libacta_ffi.so", "linux_arm64/libacta_ffi.so"],
    [linuxArchive, "x86_64-pc-windows-gnu/acta_ffi.dll", "windows_amd64/acta_ffi.dll"],
    [linuxArchive, "aarch64-pc-windows-gnullvm/acta_ffi.dll", "windows_arm64/acta_ffi.dll"],
    [macosArchive, "x86_64-apple-darwin/libacta_ffi.dylib", "darwin_amd64/libacta_ffi.dylib"],
    [macosArchive, "aarch64-apple-darwin/libacta_ffi.dylib", "darwin_arm64/libacta_ffi.dylib"],
  ]
  for (const [archive, entry, target] of mappings) {
    const stagedFile = join(staged, target)
    mkdirSync(join(stagedFile, ".."), { recursive: true })
    try {
      commandToFile("tar", ["-xzOf", archivePaths.get(archive), entry], stagedFile)
    } catch {
      annotatedError(`Required entry ${entry} is missing or invalid in ${archive}. No native files were changed.`)
    }
    if (statSync(stagedFile).size === 0) {
      annotatedError(`Required entry ${entry} is empty in ${archive}. No native files were changed.`)
    }
  }

  const nativeDirectory = join(workspace, "internal", "native")
  mkdirSync(nativeDirectory, { recursive: true })
  cpSync(staged, nativeDirectory, { recursive: true })
  writeFileSync(join(nativeDirectory, "version.go"), `package native\n\nconst EmbeddedActaVersion = "${version}"\n`)

  appendFileSync(githubOutput, `tag=${tag}\nurl=${url}\n`)
  console.log(`Fetched six FFI assets from Acta release ${tag} (${url})`)
}

main().catch((error) => {
  console.error(error.message)
  process.exitCode = 1
})
