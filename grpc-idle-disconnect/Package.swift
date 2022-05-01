// swift-tools-version: 5.6
// The swift-tools-version declares the minimum version of Swift required to build this package.

import PackageDescription

let package = Package(
    name: "swift-client",
    products: [
        // Products define the executables and libraries a package produces, and make them visible to other packages.
        .executable(
            name: "swift-client",
            targets: ["swift-client"]),
    ],
    dependencies: [
        // Dependencies declare other packages that this package depends on.
        // .package(url: /* package url */, from: "1.0.0"),
        .package(url: "https://github.com/grpc/grpc-swift.git", from: "1.0.0"),
        .package(url: "https://github.com/apple/swift-argument-parser.git", from: "1.0.0"),
    ],
    targets: [
        // Targets are the basic building blocks of a package. A target can define a module or a test suite.
        // Targets can depend on other targets in this package, and on products in packages this package depends on.
        .executableTarget(
            name: "swift-client",
            dependencies: [
                .product(name: "GRPC", package: "grpc-swift"),
                .product(name: "ArgumentParser", package: "swift-argument-parser"),
            ],
            path: ".",
            exclude: [
                "bin",
                "go-client",
                "go-server",
                "service/service.proto",
                "service/service.pb.go",
                "service/service_grpc.pb.go",
                "Makefile",
                "go.mod",
                "go.sum",
            ],
            sources: [
                "swift-client/main.swift",
                "service/service.grpc.swift",
                "service/service.pb.swift",
            ]),
    ]
)
