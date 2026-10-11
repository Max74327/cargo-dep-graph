const std = @import("std");
const config_parser = @import("config_parser.zig");

/// CLI application entry point.
pub fn main(init: std.process.Init) !void {
    const arena = init.arena.allocator();
    const io = init.io;

    var args = try init.minimal.args.iterateAllocator(arena);
    defer args.deinit();
    _ = args.next();
    const config_path: []const u8 = args.next() orelse "config.yaml";

    const cfg = config_parser.parseConfig(
        arena,
        io,
        config_path,
    ) catch |err| {
        std.debug.print(
            "Configuration error ({s}): {s}\n",
            .{ config_path, @errorName(err) },
        );
        return err;
    };

    try config_parser.printConfigToStdout(cfg, io);
}
