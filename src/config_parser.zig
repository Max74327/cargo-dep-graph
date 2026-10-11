const std = @import("std");

/// All user-configurable parameters of the application.
pub const Config = struct {
    package_name: []const u8,
    repository_url: []const u8,
    test_mode: bool,
    test_repo_path: []const u8,
    ascii_tree: bool,
    filter_substring: []const u8,
};

/// Errors that may occur while reading and parsing the configuration.
pub const ConfigError = error{
    MissingField,
    InvalidYaml,
    InvalidBool,
    EmptyPackageName,
    InvalidUrl,
    FileNotFound,
    FileTooBig,
};

/// Maximum size of the configuration file (1 MiB).
const max_file_size: std.Io.Limit = .limited(1 << 20);

/// Reads a YAML file through `io` and parses it into a Config.
/// All allocations come from `arena` and live until process exit.
pub fn parseConfig(
    arena: std.mem.Allocator,
    io: std.Io,
    path: []const u8,
) !Config {
    const cwd = std.Io.Dir.cwd();

    const content = cwd.readFileAlloc(
        io,
        path,
        arena,
        max_file_size,
    ) catch |err| switch (err) {
        error.FileNotFound => return ConfigError.FileNotFound,
        error.StreamTooLong => return ConfigError.FileTooBig,
        else => return err,
    };

    return parseConfigText(arena, content);
}

/// Parses YAML text into a Config structure.
pub fn parseConfigText(
    arena: std.mem.Allocator,
    content: []const u8,
) !Config {
    var cfg = Config{
        .package_name = "",
        .repository_url = "",
        .test_mode = false,
        .test_repo_path = "",
        .ascii_tree = false,
        .filter_substring = "",
    };

    var lines = std.mem.splitScalar(u8, content, '\n');
    while (lines.next()) |raw_line| {
        try parseLine(arena, raw_line, &cfg);
    }

    if (cfg.package_name.len == 0) return ConfigError.MissingField;
    if (cfg.repository_url.len == 0) return ConfigError.InvalidUrl;
    return cfg;
}

/// Prints all configuration parameters in key-value format.
pub fn printConfigToStdout(cfg: Config, io: std.Io) !void {
    const stdout = std.Io.File.stdout();

    try stdout.writeStreamingAll(io, "=== Configuration parameters ===\n");
    try printField(io, "package_name     ", cfg.package_name);
    try printField(io, "repository_url   ", cfg.repository_url);
    try printField(io, "test_mode        ", if (cfg.test_mode) "true" else "false");
    try printField(io, "test_repo_path   ", cfg.test_repo_path);
    try printField(io, "ascii_tree       ", if (cfg.ascii_tree) "true" else "false");
    try printField(io, "filter_substring ", cfg.filter_substring);
}

/// Writes one `key = value` line to stdout.
fn printField(io: std.Io, key: []const u8, value: []const u8) !void {
    var buf: [512]u8 = undefined;
    var w = std.Io.Writer.fixed(&buf);
    try w.print("{s} = {s}\n", .{ key, value });
    try std.Io.File.stdout().writeStreamingAll(io, w.buffered());
}

/// Parses a single YAML line and updates the corresponding Config field.
fn parseLine(
    arena: std.mem.Allocator,
    raw_line: []const u8,
    cfg: *Config,
) !void {
    const line = std.mem.trim(u8, raw_line, " \t\r");
    if (line.len == 0 or line[0] == '#') return;

    const colon = std.mem.indexOfScalar(u8, line, ':') orelse
        return ConfigError.InvalidYaml;
    const key = std.mem.trim(u8, line[0..colon], " \t");
    const value = std.mem.trim(u8, line[colon + 1 ..], " \t\"'");

    if (std.mem.eql(u8, key, "package_name")) {
        if (value.len == 0) return ConfigError.EmptyPackageName;
        cfg.package_name = try arena.dupe(u8, value);
    } else if (std.mem.eql(u8, key, "repository_url")) {
        if (value.len == 0) return ConfigError.InvalidUrl;
        cfg.repository_url = try arena.dupe(u8, value);
    } else if (std.mem.eql(u8, key, "test_mode")) {
        cfg.test_mode = try parseBool(value);
    } else if (std.mem.eql(u8, key, "test_repo_path")) {
        cfg.test_repo_path = try arena.dupe(u8, value);
    } else if (std.mem.eql(u8, key, "ascii_tree")) {
        cfg.ascii_tree = try parseBool(value);
    } else if (std.mem.eql(u8, key, "filter_substring")) {
        cfg.filter_substring = try arena.dupe(u8, value);
    }
}

/// Converts the strings "true"/"false" into a bool value.
fn parseBool(value: []const u8) !bool {
    if (std.mem.eql(u8, value, "true")) return true;
    if (std.mem.eql(u8, value, "false")) return false;
    return ConfigError.InvalidBool;
}
