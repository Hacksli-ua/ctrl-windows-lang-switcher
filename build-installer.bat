@echo off
chcp 65001 >nul
setlocal

echo ========================================
echo   Lang Switcher - Збірка інсталятора
echo ========================================
echo.

:: Шлях до Inno Setup (перевіряємо типові місця)
set "ISCC="

if exist "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" (
    set "ISCC=C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
) else if exist "C:\Program Files\Inno Setup 6\ISCC.exe" (
    set "ISCC=C:\Program Files\Inno Setup 6\ISCC.exe"
) else if exist "C:\Program Files (x86)\Inno Setup 5\ISCC.exe" (
    set "ISCC=C:\Program Files (x86)\Inno Setup 5\ISCC.exe"
) else if exist "C:\Program Files\Inno Setup 5\ISCC.exe" (
    set "ISCC=C:\Program Files\Inno Setup 5\ISCC.exe"
)

if "%ISCC%"=="" (
    echo [ПОМИЛКА] Inno Setup не знайдено!
    echo.
    echo Завантажте та встановіть Inno Setup з:
    echo https://jrsoftware.org/isdl.php
    echo.
    pause
    exit /b 1
)

echo [OK] Знайдено Inno Setup: %ISCC%
echo.

:: Перевірка наявності exe файлу
if not exist "lang-switcher.exe" (
    echo [ПОМИЛКА] lang-switcher.exe не знайдено!
    echo.
    echo Спочатку зберіть програму:
    echo   go build -ldflags="-H windowsgui" -o lang-switcher.exe
    echo.
    pause
    exit /b 1
)

echo [OK] lang-switcher.exe знайдено
echo.

:: Створення папки для інсталятора
if not exist "installer" mkdir installer

:: Збірка інсталятора
echo [*] Збірка інсталятора...
echo.

"%ISCC%" setup.iss

if %ERRORLEVEL% neq 0 (
    echo.
    echo [ПОМИЛКА] Збірка не вдалася!
    pause
    exit /b 1
)

echo.
echo ========================================
echo   Готово!
echo ========================================
echo.
echo Інсталятор: installer\LangSwitcher-Setup.exe
echo.

pause
