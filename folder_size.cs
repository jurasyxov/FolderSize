// folder_size.cs
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;

class FolderSize
{
    static string Colorize(string text, string color)
    {
        string col = color switch
        {
            "green" => "\x1b[92m",
            "red" => "\x1b[91m",
            "yellow" => "\x1b[93m",
            _ => "\x1b[0m"
        };
        return col + text + "\x1b[0m";
    }

    static string HumanReadable(long size)
    {
        string[] units = { "B", "KB", "MB", "GB", "TB" };
        double s = size;
        int i = 0;
        while (s >= 1024 && i < units.Length - 1) { s /= 1024; i++; }
        return $"{s:F1} {units[i]}";
    }

    static void GetFolderSize(string dir, bool recursive, int depth, int maxDepth,
                              bool excludeHidden, bool verbose,
                              Dictionary<string, long> sizeMap)
    {
        long total = 0;
        try
        {
            foreach (var entry in Directory.GetFileSystemEntries(dir))
            {
                string name = Path.GetFileName(entry);
                if (excludeHidden && name.StartsWith(".")) continue;
                if (File.Exists(entry))
                {
                    var info = new FileInfo(entry);
                    total += info.Length;
                    if (verbose) Console.WriteLine($"  {entry}: {HumanReadable(info.Length)}");
                }
                else if (Directory.Exists(entry) && recursive && (maxDepth == 0 || depth < maxDepth))
                {
                    var subMap = new Dictionary<string, long>();
                    GetFolderSize(entry, recursive, depth+1, maxDepth, excludeHidden, verbose, subMap);
                    long subTotal = 0;
                    foreach (var v in subMap.Values) subTotal += v;
                    total += subTotal;
                    foreach (var kv in subMap) sizeMap[kv.Key] = kv.Value;
                }
            }
        }
        catch (UnauthorizedAccessException)
        {
            if (verbose) Console.WriteLine(Colorize($"Permission denied: {dir}", "red"));
        }
        sizeMap[dir] = total;
    }

    static void Main(string[] args)
    {
        string path = ".";
        bool recursive = true, human = false, sortFlag = false, excludeHidden = false, verbose = false;
        int top = 0, maxDepth = 0;

        for (int i = 0; i < args.Length; i++)
        {
            string arg = args[i];
            if (arg == "-p" && i+1 < args.Length) path = args[++i];
            else if (arg == "-r") recursive = true;
            else if (arg == "-h") human = true;
            else if (arg == "-s") sortFlag = true;
            else if (arg == "-t" && i+1 < args.Length) top = int.Parse(args[++i]);
            else if (arg == "-d" && i+1 < args.Length) maxDepth = int.Parse(args[++i]);
            else if (arg == "--exclude-hidden") excludeHidden = true;
            else if (arg == "-v") verbose = true;
            else if (arg == "--help") { Console.WriteLine("Help..."); return; }
            else if (path == ".") path = arg;
        }

        if (!Directory.Exists(path))
        {
            Console.WriteLine(Colorize($"Error: directory '{path}' not found", "red"));
            return;
        }

        var sizeMap = new Dictionary<string, long>();
        GetFolderSize(path, recursive, 0, maxDepth, excludeHidden, verbose, sizeMap);

        var items = sizeMap.ToList();
        if (sortFlag) items.Sort((a, b) => b.Value.CompareTo(a.Value));
        if (top > 0 && top < items.Count) items = items.Take(top).ToList();

        foreach (var kv in items)
        {
            string sizeStr = human ? HumanReadable(kv.Value) : $"{kv.Value} B";
            string color = kv.Value > 1024*1024*1024 ? "red" : kv.Value > 1024*1024 ? "yellow" : "green";
            Console.WriteLine($"{Colorize(sizeStr.PadLeft(12), color)}  {kv.Key}");
        }
    }
}
