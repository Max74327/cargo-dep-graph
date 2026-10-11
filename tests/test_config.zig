const std = @import("std");
const config_parser = @import("config_parser");
const testing = std.testing;

test "parse valid config" {
    const yaml =
        \\package_name: "serde"
        \\repository_url: "https://github.com/serde-rs/serde"
        \\test_mode: true
        \\test_repo_path: "tests/fixtures/acyclic.txt"
        \\ascii_tree: false
        \\filter_substring: "tokio"
    ;

    var arena_state = std.heap.ArenaAllocator.init(testing.allocator);
    defer arena_state.deinit();
    const arena = arena_state.allocator();

    const cfg = try config_parser.parseConfigText(arena, yaml);

    try testing.expectEqualStrings("serde", cfg.package_name);
    try testing.expectEqualStrings(
        "https://github.com/serde-rs/serde",
        cfg.repository_url,
    );
    try testing.expect(cfg.test_mode);
    try testing.expect(!cfg.ascii_tree);
    try testing.expectEqualStrings("tokio", cfg.filter_substring);
}

test "reject missing package_name" {
    const yaml =
        \\repository_url: "https://github.com/serde-rs/serde"
    ;

    var arena_state = std.heap.ArenaAllocator.init(testing.allocator);
    defer arena_state.deinit();
    const arena = arena_state.allocator();

    try testing.expectError(
        config_parser.ConfigError.MissingField,
        config_parser.parseConfigText(arena, yaml),
    );
}

test "reject empty package_name" {
    const yaml =
        \\package_name: ""
        \\repository_url: "https://github.com/serde-rs/serde"
    ;

    var arena_state = std.heap.ArenaAllocator.init(testing.allocator);
    defer arena_state.deinit();
    const arena = arena_state.allocator();

    try testing.expectError(
        config_parser.ConfigError.EmptyPackageName,
        config_parser.parseConfigText(arena, yaml),
    );
}

test "reject empty repository_url" {
    const yaml =
        \\package_name: "serde"
        \\repository_url: ""
    ;

    var arena_state = std.heap.ArenaAllocator.init(testing.allocator);
    defer arena_state.deinit();
    const arena = arena_state.allocator();

    try testing.expectError(
        config_parser.ConfigError.InvalidUrl,
        config_parser.parseConfigText(arena, yaml),
    );
}

test "reject invalid bool" {
    const yaml =
        \\package_name: "serde"
        \\repository_url: "https://github.com/serde-rs/serde"
        \\test_mode: maybe
    ;

    var arena_state = std.heap.ArenaAllocator.init(testing.allocator);
    defer arena_state.deinit();
    const arena = arena_state.allocator();

    try testing.expectError(
        config_parser.ConfigError.InvalidBool,
        config_parser.parseConfigText(arena, yaml),
    );
}

test "reject line without colon" {
    const yaml =
        \\package_name: "serde"
        \\this line has no colon
    ;

    var arena_state = std.heap.ArenaAllocator.init(testing.allocator);
    defer arena_state.deinit();
    const arena = arena_state.allocator();

    try testing.expectError(
        config_parser.ConfigError.InvalidYaml,
        config_parser.parseConfigText(arena, yaml),
    );
}

test "reject missing file" {
    var arena_state = std.heap.ArenaAllocator.init(testing.allocator);
    defer arena_state.deinit();
    const arena = arena_state.allocator();

    const io = std.testing.io;

    try testing.expectError(
        config_parser.ConfigError.FileNotFound,
        config_parser.parseConfig(arena, io, "no_such_file.yaml"),
    );
}
