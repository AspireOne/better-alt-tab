param(
    [Parameter(Mandatory = $true)]
    [string]$Tag,

    [Parameter(Mandatory = $true)]
    [string]$Repo,

    [Parameter(Mandatory = $true)]
    [string]$OutputPath
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$repoUrl = "https://github.com/$Repo"
$allTags = @(git tag --sort=-creatordate)
$previousTag = $null

foreach ($existingTag in $allTags) {
    if ($existingTag -ne $Tag) {
        $previousTag = $existingTag
        break
    }
}

$lines = [System.Collections.Generic.List[string]]::new()
$lines.Add("# $Tag")
$lines.Add("")

if ($previousTag) {
    $lines.Add("## Changes Since $previousTag")
    $lines.Add("")

    $commitLines = @(git log "$previousTag..HEAD" --no-merges --pretty=format:"- %s (%h)")
    if ($commitLines.Count -eq 0) {
        $lines.Add("- No non-merge commits since the previous release.")
    }
    else {
        foreach ($commitLine in $commitLines) {
            $lines.Add($commitLine)
        }
    }

    $lines.Add("")
    $lines.Add("Compare: $repoUrl/compare/$previousTag...$Tag")
}
else {
    $lines.Add("## Initial Release")
    $lines.Add("")

    $commitLines = @(git log --no-merges --pretty=format:"- %s (%h)")
    if ($commitLines.Count -eq 0) {
        $lines.Add("- No commits found.")
    }
    else {
        foreach ($commitLine in $commitLines) {
            $lines.Add($commitLine)
        }
    }
}

[System.IO.File]::WriteAllLines($OutputPath, $lines)
