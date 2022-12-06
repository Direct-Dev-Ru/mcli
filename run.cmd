[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
@REM powershell
$env:DEBUG='true' ; go run . http -p 33333
@REM bash
DEBUG='true' go run . http -p 33333