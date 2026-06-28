// folder_size.cpp
#include <iostream>
#include <string>
#include <vector>
#include <filesystem>
#include <algorithm>
#include <iomanip>
#include <map>

using namespace std;
namespace fs = std::filesystem;

const string RESET = "\033[0m";
const string GREEN = "\033[92m";
const string RED = "\033[91m";
const string YELLOW = "\033[93m";
const string BLUE = "\033[94m";

string colorize(const string& text, const string& color) {
    return color + text + RESET;
}

string humanReadable(uintmax_t size) {
    const char* units[] = {"B", "KB", "MB", "GB", "TB"};
    double s = size;
    int i = 0;
    while (s >= 1024 && i < 4) { s /= 1024; ++i; }
    stringstream ss;
    ss << fixed << setprecision(1) << s << " " << units[i];
    return ss.str();
}

uintmax_t getFolderSize(const fs::path& dir, bool recursive, int depth, int maxDepth,
                        bool excludeHidden, bool verbose, map<string, uintmax_t>& sizeMap) {
    uintmax_t total = 0;
    try {
        for (const auto& entry : fs::directory_iterator(dir)) {
            if (excludeHidden && entry.path().filename().string()[0] == '.') continue;
            if (entry.is_regular_file()) {
                uintmax_t sz = entry.file_size();
                total += sz;
                if (verbose) {
                    cout << "  " << entry.path().string() << ": " << humanReadable(sz) << endl;
                }
            } else if (entry.is_directory() && recursive && (maxDepth == 0 || depth < maxDepth)) {
                map<string, uintmax_t> subMap;
                uintmax_t subTotal = getFolderSize(entry.path(), recursive, depth+1, maxDepth,
                                                  excludeHidden, verbose, subMap);
                total += subTotal;
                for (auto& [k, v] : subMap) sizeMap[k] = v;
            }
        }
    } catch (const exception& e) {
        if (verbose) cerr << colorize("Permission denied: " + dir.string(), RED) << endl;
    }
    sizeMap[dir.string()] = total;
    return total;
}

int main(int argc, char* argv[]) {
    string path = ".";
    bool recursive = true, human = false, sortFlag = false, excludeHidden = false, verbose = false;
    int top = 0, maxDepth = 0;

    for (int i = 1; i < argc; ++i) {
        string arg = argv[i];
        if (arg == "-p" && i+1 < argc) path = argv[++i];
        else if (arg == "-r") recursive = true;
        else if (arg == "-h") human = true;
        else if (arg == "-s") sortFlag = true;
        else if (arg == "-t" && i+1 < argc) top = stoi(argv[++i]);
        else if (arg == "-d" && i+1 < argc) maxDepth = stoi(argv[++i]);
        else if (arg == "--exclude-hidden") excludeHidden = true;
        else if (arg == "-v") verbose = true;
        else if (arg == "--help") {
            cout << "Usage: folder_size [options] [path]\n";
            return 0;
        } else if (path == ".") path = arg;
    }

    fs::path root = fs::absolute(path);
    if (!fs::exists(root) || !fs::is_directory(root)) {
        cerr << colorize("Error: invalid directory", RED) << endl;
        return 1;
    }

    map<string, uintmax_t> sizeMap;
    getFolderSize(root, recursive, 0, maxDepth, excludeHidden, verbose, sizeMap);

    vector<pair<string, uintmax_t>> items(sizeMap.begin(), sizeMap.end());
    if (sortFlag) {
        sort(items.begin(), items.end(), [](auto& a, auto& b) { return a.second > b.second; });
    }
    if (top > 0 && top < (int)items.size()) {
        items.resize(top);
    }

    for (auto& [p, size] : items) {
        string sizeStr = human ? humanReadable(size) : to_string(size) + " B";
        string color = (size > 1024*1024*1024) ? RED : (size > 1024*1024) ? YELLOW : GREEN;
        cout << colorize(string(12 - min(12, (int)sizeStr.size()), ' ') + sizeStr, color)
             << "  " << p << endl;
    }
    return 0;
}
