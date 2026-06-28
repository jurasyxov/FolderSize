// folder_size.java
import java.io.*;
import java.nio.file.*;
import java.util.*;

public class folder_size {
    private static final String RESET = "\u001B[0m";
    private static final String GREEN = "\u001B[92m";
    private static final String RED = "\u001B[91m";
    private static final String YELLOW = "\u001B[93m";

    private static String colorize(String text, String color) {
        return color + text + RESET;
    }

    private static String humanReadable(long size) {
        String[] units = {"B", "KB", "MB", "GB", "TB"};
        double s = size;
        int i = 0;
        while (s >= 1024 && i < units.length-1) { s /= 1024; i++; }
        return String.format("%.1f %s", s, units[i]);
    }

    private static void getFolderSize(Path dir, boolean recursive, int depth, int maxDepth,
                                      boolean excludeHidden, boolean verbose,
                                      Map<String, Long> sizeMap) throws IOException {
        long total = 0;
        try (DirectoryStream<Path> stream = Files.newDirectoryStream(dir)) {
            for (Path entry : stream) {
                if (excludeHidden && entry.getFileName().toString().startsWith(".")) continue;
                if (Files.isRegularFile(entry)) {
                    long sz = Files.size(entry);
                    total += sz;
                    if (verbose) System.out.println("  " + entry + ": " + humanReadable(sz));
                } else if (Files.isDirectory(entry) && recursive && (maxDepth == 0 || depth < maxDepth)) {
                    Map<String, Long> subMap = new HashMap<>();
                    getFolderSize(entry, recursive, depth+1, maxDepth, excludeHidden, verbose, subMap);
                    long subTotal = 0;
                    for (long v : subMap.values()) subTotal += v;
                    total += subTotal;
                    sizeMap.putAll(subMap);
                }
            }
        } catch (AccessDeniedException e) {
            if (verbose) System.err.println(colorize("Permission denied: " + dir, RED));
        }
        sizeMap.put(dir.toString(), total);
    }

    public static void main(String[] args) throws IOException {
        String path = ".";
        boolean recursive = true, human = false, sortFlag = false, excludeHidden = false, verbose = false;
        int top = 0, maxDepth = 0;

        for (int i = 0; i < args.length; i++) {
            String arg = args[i];
            if (arg.equals("-p") && i+1 < args.length) path = args[++i];
            else if (arg.equals("-r")) recursive = true;
            else if (arg.equals("-h")) human = true;
            else if (arg.equals("-s")) sortFlag = true;
            else if (arg.equals("-t") && i+1 < args.length) top = Integer.parseInt(args[++i]);
            else if (arg.equals("-d") && i+1 < args.length) maxDepth = Integer.parseInt(args[++i]);
            else if (arg.equals("--exclude-hidden")) excludeHidden = true;
            else if (arg.equals("-v")) verbose = true;
            else if (arg.equals("--help")) { System.out.println("Help..."); return; }
            else if (path.equals(".")) path = arg;
        }

        Path root = Paths.get(path).toAbsolutePath();
        if (!Files.isDirectory(root)) {
            System.err.println(colorize("Error: not a directory", RED));
            System.exit(1);
        }

        Map<String, Long> sizeMap = new HashMap<>();
        getFolderSize(root, recursive, 0, maxDepth, excludeHidden, verbose, sizeMap);

        List<Map.Entry<String, Long>> items = new ArrayList<>(sizeMap.entrySet());
        if (sortFlag) items.sort((a, b) -> b.getValue().compareTo(a.getValue()));
        if (top > 0 && top < items.size()) items = items.subList(0, top);

        for (var entry : items) {
            String sizeStr = human ? humanReadable(entry.getValue()) : entry.getValue() + " B";
            String color = entry.getValue() > 1024*1024*1024 ? RED :
                           entry.getValue() > 1024*1024 ? YELLOW : GREEN;
            System.out.println(colorize(String.format("%12s", sizeStr), color) + "  " + entry.getKey());
        }
    }
}
