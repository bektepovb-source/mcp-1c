@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:: Собирает .cfe расширения MCP_HTTPService из XML-исходников.
:: Требует установленной 1C:Предприятие (учебная или полная версия).
::
:: Использование:
::   scripts\build-extension.cmd C:\Users\User\Documents\InfoBase
::   scripts\build-extension.cmd C:\Users\User\Documents\InfoBase C:\out\MCP_HTTPService.cfe
::
:: Можно задать бинарник явно:
::   set DESIGNER=C:\...\1cv8.exe
::   scripts\build-extension.cmd C:\Users\User\Documents\InfoBase

set "SCRIPT_DIR=%~dp0"
set "PROJECT_DIR=%SCRIPT_DIR%.."
set "EXTENSION_SRC=%PROJECT_DIR%\extension\src"
set "EXTENSION_NAME=MCP_HTTPService"

:: ── Аргументы ──────────────────────────────────────
set "INFOBASE=%~1"
set "OUTPUT=%~2"

if "%INFOBASE%"=="" (
    echo Использование: scripts\build-extension.cmd ^<путь_к_базе^> [путь_к_output.cfe]
    echo Пример:        scripts\build-extension.cmd C:\Users\User\Documents\InfoBase
    exit /b 1
)

if "%OUTPUT%"=="" set "OUTPUT=%PROJECT_DIR%\extension\%EXTENSION_NAME%.cfe"

:: ── Проверяем исходники ────────────────────────────
if not exist "%EXTENSION_SRC%\Configuration.xml" (
    echo Ошибка: не найден %EXTENSION_SRC%\Configuration.xml
    exit /b 1
)

:: ── Если DESIGNER задан явно — используем его ──────
if defined DESIGNER (
    if not exist "%DESIGNER%" (
        echo Ошибка: указанный DESIGNER не найден: %DESIGNER%
        exit /b 1
    )
    goto :designer_found
)

:: ── Поиск всех версий 1C ──────────────────────────
set "COUNT=0"

for /d %%d in ("C:\Program Files\1cv8\8.*") do (
    if exist "%%d\bin\1cv8.exe" (
        set /a COUNT+=1
        set "BIN_!COUNT!=%%d\bin\1cv8.exe"
        set "LABEL_!COUNT!=коммерческая %%d"
    )
)
for /d %%d in ("C:\Program Files (x86)\1cv8\8.*") do (
    if exist "%%d\bin\1cv8.exe" (
        set /a COUNT+=1
        set "BIN_!COUNT!=%%d\bin\1cv8.exe"
        set "LABEL_!COUNT!=коммерческая %%d"
    )
)
for /d %%d in ("C:\Program Files\1cv8t\8.*") do (
    if exist "%%d\bin\1cv8t.exe" (
        set /a COUNT+=1
        set "BIN_!COUNT!=%%d\bin\1cv8t.exe"
        set "LABEL_!COUNT!=учебная %%d"
    )
)
for /d %%d in ("C:\Program Files (x86)\1cv8t\8.*") do (
    if exist "%%d\bin\1cv8t.exe" (
        set /a COUNT+=1
        set "BIN_!COUNT!=%%d\bin\1cv8t.exe"
        set "LABEL_!COUNT!=учебная %%d"
    )
)
for /d %%d in ("C:\1cv8t\8.*") do (
    if exist "%%d\bin\1cv8t.exe" (
        set /a COUNT+=1
        set "BIN_!COUNT!=%%d\bin\1cv8t.exe"
        set "LABEL_!COUNT!=учебная %%d"
    )
)

if %COUNT% equ 0 (
    echo Ошибка: не найден бинарник 1C. Установите 1C:Предприятие.
    echo Или задайте: set DESIGNER=C:\...\1cv8.exe
    exit /b 1
)

if %COUNT% equ 1 (
    set "DESIGNER=!BIN_1!"
    goto :designer_found
)

:: Несколько версий — спрашиваем
echo Найдено несколько версий 1C:
for /l %%i in (1,1,%COUNT%) do (
    echo   %%i^) !LABEL_%%i!
)
set /p "CHOICE=Выберите версию (1-%COUNT%): "

if not defined CHOICE goto :bad_choice
if %CHOICE% lss 1 goto :bad_choice
if %CHOICE% gtr %COUNT% goto :bad_choice

set "DESIGNER=!BIN_%CHOICE%!"
goto :designer_found

:bad_choice
echo Ошибка: неверный выбор
exit /b 1

:designer_found
echo 1C: %DESIGNER%

:: ── Загрузка XML ───────────────────────────────────
echo Загружаем XML в расширение %EXTENSION_NAME%...
"%DESIGNER%" DESIGNER /F "%INFOBASE%" /LoadConfigFromFiles "%EXTENSION_SRC%" -Extension "%EXTENSION_NAME%"

if %errorlevel% neq 0 (
    echo Ошибка: LoadConfigFromFiles завершилась с кодом %errorlevel%
    exit /b %errorlevel%
)

:: ── Выгрузка .cfe ──────────────────────────────────
for %%F in ("%OUTPUT%") do if not exist "%%~dpF" mkdir "%%~dpF"
echo Выгружаем .cfe...
"%DESIGNER%" DESIGNER /F "%INFOBASE%" /DumpCfg "%OUTPUT%" -Extension "%EXTENSION_NAME%"

if %errorlevel% neq 0 (
    echo Ошибка: DumpCfg завершилась с кодом %errorlevel%
    exit /b %errorlevel%
)

echo Готово: %OUTPUT%
dir "%OUTPUT%"
