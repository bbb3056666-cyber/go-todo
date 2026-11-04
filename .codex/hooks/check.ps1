$ErrorActionPreference = "Stop"

$eventData = [Console]::In.ReadToEnd() | ConvertFrom-Json

# Only send feedback once, avoiding an endless continuation loop.
if ($eventData.stop_hook_active) {
    "{}"
    exit 0
}

try {
    Set-Location F:\go-todo

    # Merge output inside cmd, avoiding PowerShell stderr handling differences.
    $checkOutput = & cmd.exe /d /c "make check 2>&1"
    $checkExitCode = $LASTEXITCODE

    # Limit the feedback size.
    $details = ($checkOutput | Select-Object -Last 80) -join "`n"

    if ($checkExitCode -eq 0) {
        $reason = @"
make check PASSED.
Report this verification result to the user.

Output:
$details
"@
    } else {
        $reason = @"
make check FAILED with exit code $checkExitCode.
Inspect the output. Fix issues only within the user's authorized scope,
then rerun make check using your own command tool.
If blocked, report the actual blocker. Do not claim checks passed.

Output:
$details
"@
    }
}
catch {
    $reason = @"
The verification hook could not run make check:
$($_.Exception.Message)
Investigate and report the actual verification status.
"@
}

[Console]::Error.WriteLine($reason)
exit 2