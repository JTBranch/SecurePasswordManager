@echo off
REM Go Password Manager Installer for Windows

echo 🔐 Go Password Manager Installer
echo ================================
echo.

REM Create installation directory
set INSTALL_DIR=%USERPROFILE%\AppData\Local\PasswordManager
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

echo 📥 Downloading latest version...

REM Use PowerShell to download (more reliable than curl on Windows)
powershell -Command ^
    "$latest = Invoke-RestMethod 'https://api.github.com/repos/JTBranch/SecurePasswordManager/releases/latest'; ^
     $asset = $latest.assets | Where-Object { $_.name -eq 'password-manager-windows-amd64.exe' }; ^
     if ($asset) { ^
         Write-Host '📦 Installing version' $latest.tag_name; ^
         Invoke-WebRequest -Uri $asset.browser_download_url -OutFile '%INSTALL_DIR%\password-manager.exe'; ^
         Write-Host '✅ Download complete'; ^
     } else { ^
         Write-Host '❌ Failed to find Windows binary'; ^
         exit 1; ^
     }"

if %errorlevel% neq 0 (
    echo ❌ Download failed
    pause
    exit /b 1
)

REM Create desktop shortcut
set DESKTOP=%USERPROFILE%\Desktop
echo Set oWS = WScript.CreateObject("WScript.Shell") > "%TEMP%\createShortcut.vbs"
echo sLinkFile = "%DESKTOP%\Password Manager.lnk" >> "%TEMP%\createShortcut.vbs"
echo Set oLink = oWS.CreateShortcut(sLinkFile) >> "%TEMP%\createShortcut.vbs"
echo oLink.TargetPath = "%INSTALL_DIR%\password-manager.exe" >> "%TEMP%\createShortcut.vbs"
echo oLink.WorkingDirectory = "%INSTALL_DIR%" >> "%TEMP%\createShortcut.vbs"
echo oLink.Description = "Go Password Manager" >> "%TEMP%\createShortcut.vbs"
echo oLink.Save >> "%TEMP%\createShortcut.vbs"
cscript /nologo "%TEMP%\createShortcut.vbs"
del "%TEMP%\createShortcut.vbs"

echo.
echo 🎉 Installation Complete!
echo.
echo 📂 Installed to: %INSTALL_DIR%
echo 🖥️  Desktop shortcut: Password Manager
echo.
echo 🚀 To run the app:
echo    1. Double-click 'Password Manager' on your Desktop
echo    2. Or run: %INSTALL_DIR%\password-manager.exe
echo.
echo 📋 Your passwords will be stored in: %INSTALL_DIR%\secrets.json
echo.
echo Thank you for using Go Password Manager! 🔐
echo.
pause
