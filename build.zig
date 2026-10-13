const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const config_parser = b.addLibrary(.{
        .name = "config_parser",
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/config_parser.zig"),
            .target = target,
            .optimize = optimize,
        }),
        .linkage = .static,
    });

    const docs_index = b.graph.addGeneratedFile(&config_parser.step);
    config_parser.generated_docs = .init(docs_index);

    b.installArtifact(config_parser);

    const exe = b.addExecutable(.{
        .name = "cargo-dep-graph",
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/main.zig"),
            .target = target,
            .optimize = optimize,
        }),
    });

    exe.root_module.addImport("config_parser", config_parser.root_module);
    b.installArtifact(exe);

    const run_cmd = b.addRunArtifact(exe);
    run_cmd.step.dependOn(b.getInstallStep());
    const run_step = b.step("run", "Run the application");
    run_step.dependOn(&run_cmd.step);

    const unit_tests = b.addTest(.{
        .root_module = b.createModule(.{
            .root_source_file = b.path("tests/test_config.zig"),
            .target = target,
            .optimize = optimize,
        }),
    });
    unit_tests.root_module.addImport(
        "config_parser",
        config_parser.root_module,
    );

    const run_unit_tests = b.addRunArtifact(unit_tests);
    const test_step = b.step("test", "Run unit tests");
    test_step.dependOn(&run_unit_tests.step);

    const install_docs = b.addInstallDirectory(.{
        .source_dir = config_parser.getEmittedDocs(),
        .install_dir = .prefix,
        .install_subdir = "docs",
    });

    const docs_step = b.step("docs", "Generate HTML documentation");
    docs_step.dependOn(&install_docs.step);
}
