# folder_size.py
#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys
import argparse
from pathlib import Path
from collections import defaultdict

# ANSI colors
COLORS = {
    'green': '\033[92m',
    'red': '\033[91m',
    'yellow': '\033[93m',
    'blue': '\033[94m',
    'reset': '\033[0m'
}

def colorize(text, color):
    return f"{COLORS.get(color, '')}{text}{COLORS['reset']}"

def human_readable(size, decimal_places=1):
    for unit in ['B', 'KB', 'MB', 'GB', 'TB']:
        if size < 1024.0:
            return f"{size:{decimal_places+1}.{decimal_places}f} {unit}"
        size /= 1024.0
    return f"{size:.1f} PB"

def get_folder_size(path, recursive=True, depth=0, max_depth=None, exclude_hidden=False, verbose=False):
    """Рекурсивно вычисляет размер папки и возвращает словарь {путь: размер}."""
    path = Path(path)
    if not path.exists():
        raise FileNotFoundError(f"Path not found: {path}")
    if not path.is_dir():
        raise NotADirectoryError(f"Not a directory: {path}")

    result = {}
    total_size = 0

    try:
        items = list(path.iterdir())
    except PermissionError:
        if verbose:
            print(colorize(f"Permission denied: {path}", 'red'))
        return {str(path): 0}

    for item in items:
        if exclude_hidden and item.name.startswith('.'):
            continue
        if item.is_file():
            try:
                size = item.stat().st_size
                total_size += size
                if verbose:
                    print(f"  {item}: {human_readable(size)}")
            except OSError:
                pass
        elif item.is_dir() and recursive and (max_depth is None or depth < max_depth):
            sub_result = get_folder_size(item, recursive, depth+1, max_depth, exclude_hidden, verbose)
            if sub_result:
                sub_size = sum(sub_result.values())
                total_size += sub_size
                result.update(sub_result)
        elif item.is_dir() and not recursive:
            # Если не рекурсивно, просто пропускаем папки
            pass

    result[str(path)] = total_size
    return result

def main():
    parser = argparse.ArgumentParser(description="FolderSize – измерение размера папок")
    parser.add_argument('path', nargs='?', default='.', help='Путь к папке (по умолчанию текущая)')
    parser.add_argument('-r', '--recursive', action='store_true', default=True, help='Рекурсивно (включено)')
    parser.add_argument('-h', '--human-readable', action='store_true', help='Выводить размер в удобном формате')
    parser.add_argument('-s', '--sort', action='store_true', help='Сортировать по размеру (убывание)')
    parser.add_argument('-t', '--top', type=int, help='Показать только N самых больших элементов')
    parser.add_argument('-d', '--depth', type=int, help='Максимальная глубина рекурсии (0 – только текущая)')
    parser.add_argument('--exclude-hidden', action='store_true', help='Игнорировать скрытые файлы и папки')
    parser.add_argument('-v', '--verbose', action='store_true', help='Подробный вывод')
    args = parser.parse_args()

    try:
        size_map = get_folder_size(args.path, args.recursive, 0, args.depth,
                                   args.exclude_hidden, args.verbose)
    except Exception as e:
        sys.exit(colorize(f"Ошибка: {e}", 'red'))

    if not size_map:
        print(colorize("Нет данных.", 'yellow'))
        return

    # Преобразуем в список (путь, размер)
    items = list(size_map.items())
    if args.sort:
        items.sort(key=lambda x: x[1], reverse=True)
    if args.top:
        items = items[:args.top]

    # Определяем максимальную длину пути для выравнивания
    max_path_len = max(len(p) for p, _ in items) if items else 0

    for path, size in items:
        if args.human_readable:
            size_str = human_readable(size)
        else:
            size_str = f"{size:,} B"
        # Цвет для размера
        if size < 1024 * 1024:
            color = 'green'
        elif size < 1024 * 1024 * 1024:
            color = 'yellow'
        else:
            color = 'red'
        print(f"{colorize(size_str.rjust(12), color)}  {path}")

if __name__ == '__main__':
    main()
