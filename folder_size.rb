#!/usr/bin/env ruby
# folder_size.rb
# encoding: UTF-8

require 'find'
require 'optparse'
require 'pathname'

COLORS = {
  green: "\e[92m",
  red: "\e[91m",
  yellow: "\e[93m",
  reset: "\e[0m"
}

def colorize(text, color)
  "#{COLORS[color]}#{text}#{COLORS[:reset]}"
end

def human_readable(size)
  units = ['B', 'KB', 'MB', 'GB', 'TB']
  s = size.to_f
  i = 0
  while s >= 1024 && i < units.length - 1
    s /= 1024
    i += 1
  end
  sprintf("%.1f %s", s, units[i])
end

def get_folder_size(dir, recursive, depth, max_depth, exclude_hidden, verbose, size_map)
  total = 0
  begin
    entries = Dir.entries(dir) - ['.', '..']
  rescue Errno::EACCES
    puts colorize("Permission denied: #{dir}", :red) if verbose
    size_map[dir] = 0
    return 0
  end

  entries.each do |name|
    next if exclude_hidden && name.start_with?('.')
    full = File.join(dir, name)
    if File.file?(full)
      sz = File.size(full)
      total += sz
      puts "  #{full}: #{human_readable(sz)}" if verbose
    elsif File.directory?(full) && recursive && (max_depth == 0 || depth < max_depth)
      sub_map = {}
      sub_total = get_folder_size(full, recursive, depth+1, max_depth, exclude_hidden, verbose, sub_map)
      total += sub_total
      sub_map.each { |k, v| size_map[k] = v }
    end
  end
  size_map[dir] = total
  total
end

options = {
  path: '.',
  recursive: true,
  human: false,
  sort: false,
  top: 0,
  depth: 0,
  exclude_hidden: false,
  verbose: false
}

OptionParser.new do |opts|
  opts.banner = "Usage: folder_size.rb [options] [path]"
  opts.on("-p", "--path DIR", "Path") { |v| options[:path] = v }
  opts.on("-r", "--recursive", "Recursive") { options[:recursive] = true }
  opts.on("-h", "--human-readable", "Human readable") { options[:human] = true }
  opts.on("-s", "--sort", "Sort by size") { options[:sort] = true }
  opts.on("-t", "--top N", Integer, "Show top N") { |v| options[:top] = v }
  opts.on("-d", "--depth N", Integer, "Max depth") { |v| options[:depth] = v }
  opts.on("--exclude-hidden", "Exclude hidden") { options[:exclude_hidden] = true }
  opts.on("-v", "--verbose", "Verbose") { options[:verbose] = true }
  opts.on("-h", "--help", "Help") { puts opts; exit }
end.parse!

root = File.expand_path(options[:path] || ARGV[0] || '.')
unless Dir.exist?(root)
  puts colorize("Error: directory not found", :red)
  exit 1
end

size_map = {}
get_folder_size(root, options[:recursive], 0, options[:depth], options[:exclude_hidden],
                options[:verbose], size_map)

items = size_map.to_a
items.sort! { |a, b| b[1] <=> a[1] } if options[:sort]
if options[:top] > 0 && options[:top] < items.size
  items = items[0...options[:top]]
end

max_path_len = items.map { |p, _| p.length }.max || 0
items.each do |path, size|
  size_str = options[:human] ? human_readable(size) : "#{size} B"
  color = if size > 1024*1024*1024
            :red
          elsif size > 1024*1024
            :yellow
          else
            :green
          end
  puts "#{colorize(size_str.rjust(12), color)}  #{path}"
end
