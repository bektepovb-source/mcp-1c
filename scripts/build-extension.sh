#!/bin/bash
# Собирает .cfe расширения MCP_HTTPService из XML-исходников.
# Требует установленной 1C:Предприятие (учебная или полная версия).
#
# Использование:
#   ./scripts/build-extension.sh /path/to/infobase [/path/to/output.cfe]
#
# Можно задать бинарник явно:
#   DESIGNER=/path/to/1cv8 ./scripts/build-extension.sh ~/Documents/InfoBase

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
EXTENSION_SRC="$PROJECT_DIR/extension/src"
EXTENSION_NAME="MCP_HTTPService"

# Собираем все найденные бинарники 1C (macOS)
collect_1c_binaries() {
    local bins=()
    for d in /Applications/1cv8t.localized/*/1cv8t.app/Contents/MacOS/1cv8t; do
        [ -f "$d" ] && bins+=("$d")
    done
    for d in /Applications/1cv8.localized/*/1cv8.app/Contents/MacOS/1cv8; do
        [ -f "$d" ] && bins+=("$d")
    done
    printf '%s\n' "${bins[@]}"
}

# Аргументы
INFOBASE="${1:?Использование: ./scripts/build-extension.sh <путь_к_базе> [путь_к_output.cfe]}"
OUTPUT="${2:-$PROJECT_DIR/extension/$EXTENSION_NAME.cfe}"

# Если DESIGNER задан явно — используем его
if [ -n "${DESIGNER:-}" ]; then
    if [ ! -f "$DESIGNER" ]; then
        echo "Ошибка: указанный DESIGNER не найден: $DESIGNER" >&2
        exit 1
    fi
else
    # Ищем все доступные версии
    FOUND=()
    while IFS= read -r line; do
        [ -n "$line" ] && FOUND+=("$line")
    done < <(collect_1c_binaries)

    if [ ${#FOUND[@]} -eq 0 ]; then
        echo "Ошибка: не найден бинарник 1C. Установите 1C:Предприятие." >&2
        echo "Или задайте: DESIGNER=/path/to/1cv8 ./scripts/build-extension.sh ..." >&2
        exit 1
    elif [ ${#FOUND[@]} -eq 1 ]; then
        DESIGNER="${FOUND[0]}"
    else
        echo "Найдено несколько версий 1C:"
        for i in "${!FOUND[@]}"; do
            echo "  $((i + 1))) ${FOUND[$i]}"
        done
        read -rp "Выберите версию (1-${#FOUND[@]}): " CHOICE
        if ! [[ "$CHOICE" =~ ^[0-9]+$ ]] || [ "$CHOICE" -lt 1 ] || [ "$CHOICE" -gt ${#FOUND[@]} ]; then
            echo "Ошибка: неверный выбор" >&2
            exit 1
        fi
        DESIGNER="${FOUND[$((CHOICE - 1))]}"
    fi
fi

echo "1C: $DESIGNER"

# Проверяем исходники
if [ ! -f "$EXTENSION_SRC/Configuration.xml" ]; then
    echo "Ошибка: не найден $EXTENSION_SRC/Configuration.xml" >&2
    exit 1
fi

# Загрузка XML в расширение
echo "Загружаем XML в расширение $EXTENSION_NAME..."
"$DESIGNER" DESIGNER \
    /F "$INFOBASE" \
    /LoadConfigFromFiles "$EXTENSION_SRC" \
    -Extension "$EXTENSION_NAME" || {
    echo "Ошибка: LoadConfigFromFiles завершилась с ошибкой" >&2
    exit 1
}

# Выгрузка .cfe
echo "Выгружаем .cfe..."
mkdir -p "$(dirname "$OUTPUT")"
"$DESIGNER" DESIGNER \
    /F "$INFOBASE" \
    /DumpCfg "$OUTPUT" \
    -Extension "$EXTENSION_NAME" || {
    echo "Ошибка: DumpCfg завершилась с ошибкой" >&2
    exit 1
}

echo "Готово: $OUTPUT"
ls -lh "$OUTPUT"
